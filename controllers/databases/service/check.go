package service

import (
	"errors"
	"fmt"
	"net"
	"time"

	databasesv1 "github.com/DevineLiu/redis-operator/apis/databases/v1"
	"github.com/DevineLiu/redis-operator/controllers/databases/util"
	util2 "github.com/DevineLiu/redis-operator/controllers/databases/util"
	"github.com/DevineLiu/redis-operator/controllers/middle/client/k8s"
	"github.com/DevineLiu/redis-operator/controllers/middle/client/redis"
	"github.com/go-logr/logr"
	goredis "github.com/go-redis/redis"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type RedisFailoverCheck interface {
	CheckRedisNumber(rf *databasesv1.RedisFailover) error
	CheckSentinelNumber(rf *databasesv1.RedisFailover) error
	CheckSentinelReadyReplicas(rf *databasesv1.RedisFailover) error
	CheckAllSlavesFromMaster(master string, rf *databasesv1.RedisFailover, auth *util2.AuthConfig) error
	CheckSentinelNumberInMemory(sentinel string, rf *databasesv1.RedisFailover, auth *util2.AuthConfig) error
	CheckSentinelSlavesNumberInMemory(sentinel string, rf *databasesv1.RedisFailover, auth *util2.AuthConfig) error
	CheckSentinelMonitor(sentinel string, monitor string, auth *util2.AuthConfig) error
	GetMasterIP(rf *databasesv1.RedisFailover, auth *util2.AuthConfig) (string, error)
	GetNumberMasters(rf *databasesv1.RedisFailover, auth *util2.AuthConfig) (int, error)
	GetRedisesIPs(rf *databasesv1.RedisFailover, auth *util2.AuthConfig) ([]string, error)
	GetSentinelsIPs(rf *databasesv1.RedisFailover) ([]string, error)
	GetMinimumRedisPodTime(rf *databasesv1.RedisFailover) (time.Duration, error)
	CheckRedisConfig(rf *databasesv1.RedisFailover, addr string, auth *util2.AuthConfig) error
}

type RedisFailoverChecker struct {
	K8SService   k8s.Services
	Logger       logr.Logger
	StatusWriter client.StatusWriter
	Record       record.EventRecorder
	RedisClient  redis.Client
}

func NewRedisFailoverChecker(k8SService k8s.Services, log logr.Logger, status client.StatusWriter, record record.EventRecorder, rc redis.Client) *RedisFailoverChecker {
	return &RedisFailoverChecker{
		K8SService: k8SService, Logger: log, StatusWriter: status, Record: record, RedisClient: rc,
	}
}

func (r RedisFailoverChecker) CheckRedisNumber(rf *databasesv1.RedisFailover) error {
	ss, err := r.K8SService.GetStatefulSet(rf.Namespace, util2.GetRedisName(rf))
	if err != nil {
		return err
	}
	if rf.Spec.Redis.Replicas != *ss.Spec.Replicas {

		return fmt.Errorf("number  of stateful differ from spec,cr: %d ss: %d", rf.Spec.Redis.Replicas, ss.Spec.Replicas)
	}
	if rf.Spec.Redis.Replicas != ss.Status.ReadyReplicas {
		return fmt.Errorf("waiting all of redis pods become ready,cr:%d ss ready:%d", rf.Spec.Redis.Replicas, ss.Status.ReadyReplicas)
	}
	return nil
}

func (r RedisFailoverChecker) CheckSentinelNumber(rf *databasesv1.RedisFailover) error {
	deploy, err := r.K8SService.GetDeployment(rf.Namespace, util2.GetSentinelName(rf))
	if err != nil {
		return err
	}
	if rf.Spec.Sentinel.Replicas != *deploy.Spec.Replicas {
		return fmt.Errorf("number of sentinel pods differ from spec,cr:%d deploy:%d", rf.Spec.Sentinel.Replicas, deploy.Spec.Replicas)
	}
	return err
}

func (r RedisFailoverChecker) CheckSentinelReadyReplicas(rf *databasesv1.RedisFailover) error {
	d, err := r.K8SService.GetDeployment(rf.Namespace, util.GetSentinelName(rf))
	if err != nil {
		return err
	}
	if rf.Spec.Sentinel.Replicas != d.Status.ReadyReplicas {
		return errors.New("waiting all of sentinel pods become ready")
	}
	return nil
}

func (r RedisFailoverChecker) CheckAllSlavesFromMaster(master string, rf *databasesv1.RedisFailover, auth *util2.AuthConfig) error {
	rips, err := r.GetRedisesIPs(rf, auth)
	if err != nil {
		return err
	}
	for _, rip := range rips {
		slave, err := r.RedisClient.GetSlaveMasterIP(rip, auth)
		if err != nil {
			return err
		}
		if slave != "" && slave != master {
			return fmt.Errorf("slave %s don't have the master %s, has %s", rip, master, slave)
		}
	}
	return nil
}

func (r RedisFailoverChecker) CheckSentinelNumberInMemory(sentinel string, rf *databasesv1.RedisFailover, auth *util2.AuthConfig) error {
	nSentinels, err := r.RedisClient.GetNumberSentinelsInMemory(sentinel, auth)
	if err != nil {
		return err
	} else if nSentinels != rf.Spec.Sentinel.Replicas {
		return errors.New("sentinels in memory mismatch")
	}
	return nil
}

func (r RedisFailoverChecker) CheckSentinelSlavesNumberInMemory(sentinel string, rf *databasesv1.RedisFailover, auth *util2.AuthConfig) error {
	nSlaves, err := r.RedisClient.GetNumberSentinelSlavesInMemory(sentinel, auth)
	if err != nil {
		return err
	} else if nSlaves != rf.Spec.Redis.Replicas-1 {
		return errors.New("sentinel's slaves in memory mismatch")
	}
	return nil
}

func (r RedisFailoverChecker) CheckSentinelMonitor(sentinel string, monitor string, auth *util2.AuthConfig) error {
	actualMonitorIP, err := r.RedisClient.GetSentinelMonitor(sentinel, auth)
	if err != nil {
		return err
	}
	if actualMonitorIP != monitor {
		return errors.New("the monitor on the sentinel config does not match with the expected one")
	}
	return nil
}

func (r RedisFailoverChecker) GetMasterIP(rf *databasesv1.RedisFailover, auth *util2.AuthConfig) (string, error) {
	rips, err := r.GetRedisesIPs(rf, auth)
	if err != nil {
		return "", err
	}
	masters := []string{}
	for _, rip := range rips {
		master, err := r.RedisClient.IsMaster(rip, auth)
		if err != nil {
			return "", err
		}
		if master {
			masters = append(masters, rip)
		}
	}

	if len(masters) != 1 {
		return "", errors.New("number of redis nodes known as master is different than 1")
	}
	return masters[0], nil
}

func (r RedisFailoverChecker) GetNumberMasters(rf *databasesv1.RedisFailover, auth *util2.AuthConfig) (int, error) {
	nMasters := 0
	rips, err := r.GetRedisesIPs(rf, auth)
	if err != nil {
		return nMasters, err
	}
	for _, rip := range rips {
		master, err := r.RedisClient.IsMaster(rip, auth)
		if err != nil {
			return nMasters, err
		}
		if master {
			nMasters++
		}
	}
	return nMasters, err
}

func (r RedisFailoverChecker) GetRedisesIPs(rf *databasesv1.RedisFailover, auth *util2.AuthConfig) ([]string, error) {
	redisips := []string{}
	rps, err := r.K8SService.GetStatefulSetPods(rf.Namespace, util2.GetRedisName(rf))
	if err != nil {
		return nil, err
	}
	for _, rp := range rps.Items {
		if rp.Status.Phase == corev1.PodRunning { // Only work with running pods
			redisips = append(redisips, rp.Status.PodIP)
		}
	}
	return redisips, nil
}

func (r RedisFailoverChecker) GetSentinelsIPs(rf *databasesv1.RedisFailover) ([]string, error) {
	sentinels := []string{}
	rps, err := r.K8SService.GetDeploymentPods(rf.Namespace, util2.GetSentinelName(rf))
	if err != nil {
		return nil, err
	}
	for _, sp := range rps.Items {
		if sp.Status.Phase == corev1.PodRunning { // Only work with running pods
			sentinels = append(sentinels, sp.Status.PodIP)
		}
	}
	return sentinels, nil
}

func (r RedisFailoverChecker) GetMinimumRedisPodTime(rf *databasesv1.RedisFailover) (time.Duration, error) {
	minTime := 100000 * time.Hour // More than ten years
	rps, err := r.K8SService.GetStatefulSetPods(rf.Namespace, util2.GetRedisName(rf))
	if err != nil {
		return minTime, err
	}
	for _, redisNode := range rps.Items {
		if redisNode.Status.StartTime == nil {
			continue
		}
		start := redisNode.Status.StartTime.Round(time.Second)
		alive := time.Now().Sub(start)
		if alive < minTime {
			minTime = alive
		}
	}
	return minTime, nil
}

func (r RedisFailoverChecker) CheckRedisConfig(rf *databasesv1.RedisFailover, addr string, auth *util2.AuthConfig) error {
	client := goredis.NewClient(&goredis.Options{
		Addr:     net.JoinHostPort(addr, "6379"),
		Password: auth.Password,
		DB:       0,
	})
	defer client.Close()
	configs, err := r.RedisClient.GetAllRedisConfig(client)
	if err != nil {
		return err
	}

	for key, value := range rf.Spec.Redis.CustomConfig {
		var err error
		if _, ok := parseConfigMap[key]; ok {
			value, err = util.ParseRedisMemConf(value)
			if err != nil {
				r.Logger.Error(err, "redis  format err", "key", key, "value", value)
				continue
			}
		}
		if value != configs[key] {
			return fmt.Errorf("%s configs conflict, expect: %s, current: %s", key, value, configs[key])
		}
	}
	return nil
}

var parseConfigMap = map[string]int8{
	"maxmemory":                  0,
	"proto-max-bulk-len":         0,
	"client-query-buffer-limit":  0,
	"repl-backlog-size":          0,
	"auto-aof-rewrite-min-size":  0,
	"active-defrag-ignore-bytes": 0,
	"hash-max-ziplist-entries":   0,
	"hash-max-ziplist-value":     0,
	"stream-node-max-bytes":      0,
	"set-max-intset-entries":     0,
	"zset-max-ziplist-entries":   0,
	"zset-max-ziplist-value":     0,
	"hll-sparse-max-bytes":       0,
	// TODO parse client-output-buffer-limit
	//"client-output-buffer-limit": 0,
}
