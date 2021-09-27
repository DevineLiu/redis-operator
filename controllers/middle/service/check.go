package service

import (
	"errors"
	"fmt"
	"github.com/DevineLiu/redis-operator/apis/middle/v1alpha1"
	"github.com/DevineLiu/redis-operator/controllers/middle/client/k8s"
	"github.com/DevineLiu/redis-operator/controllers/middle/client/redis"
	"github.com/DevineLiu/redis-operator/util"
	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"time"
)

type RedisFailoverCheck interface {
	CheckRedisNumber(rf *v1alpha1.RedisFailover) error
	CheckSentinelNumber(rf *v1alpha1.RedisFailover) error
	CheckSentinelReadyReplicas(rf *v1alpha1.RedisFailover) error
	CheckAllSlavesFromMaster(master string, rf *v1alpha1.RedisFailover, auth *util.AuthConfig) error
	CheckSentinelNumberInMemory(sentinel string, rf *v1alpha1.RedisFailover, auth *util.AuthConfig) error
	CheckSentinelSlavesNumberInMemory(sentinel string, rf *v1alpha1.RedisFailover, auth *util.AuthConfig) error
	CheckSentinelMonitor(sentinel string, monitor string, auth *util.AuthConfig) error
	GetMasterIP(rf *v1alpha1.RedisFailover, auth *util.AuthConfig) (string, error)
	GetNumberMasters(rf *v1alpha1.RedisFailover, auth *util.AuthConfig) (int, error)
	GetRedisesIPs(rf *v1alpha1.RedisFailover, auth *util.AuthConfig) ([]string, error)
	GetSentinelsIPs(rf *v1alpha1.RedisFailover) ([]string, error)
	GetMinimumRedisPodTime(rf *v1alpha1.RedisFailover) (time.Duration, error)
	CheckRedisConfig(rf *v1alpha1.RedisFailover, addr string, auth *util.AuthConfig) error
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

func (r RedisFailoverChecker) CheckRedisNumber(rf *v1alpha1.RedisFailover) error {
	ss, err := r.K8SService.GetStatefulSet(rf.Namespace, util.GetRedisName(rf))
	if err != nil {
		return err
	}
	if rf.Spec.Redis.Replicas != *ss.Spec.Replicas {
		return errors.New("number  of stateful differ from spec")
	}
	if rf.Spec.Redis.Replicas != ss.Status.ReadyReplicas {
		return errors.New("waiting all of redis pods become ready")
	}
	return nil
}

func (r RedisFailoverChecker) CheckSentinelNumber(rf *v1alpha1.RedisFailover) error {
	deploy, err := r.K8SService.GetDeployment(rf.Namespace, util.GetSentinelName(rf))
	if err != nil {
		return err
	}
	if rf.Spec.Redis.Replicas != *deploy.Spec.Replicas {
		return errors.New("number of sentinel pos differ from spec")
	}
	return err
}

func (r RedisFailoverChecker) CheckSentinelReadyReplicas(rf *v1alpha1.RedisFailover) error {
	panic("implement me")
}

func (r RedisFailoverChecker) CheckAllSlavesFromMaster(master string, rf *v1alpha1.RedisFailover, auth *util.AuthConfig) error {
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

func (r RedisFailoverChecker) CheckSentinelNumberInMemory(sentinel string, rf *v1alpha1.RedisFailover, auth *util.AuthConfig) error {
	nSentinels, err := r.RedisClient.GetNumberSentinelsInMemory(sentinel, auth)
	if err != nil {
		return err
	} else if nSentinels != rf.Spec.Sentinel.Replicas {
		return errors.New("sentinels in memory mismatch")
	}
	return nil
}

func (r RedisFailoverChecker) CheckSentinelSlavesNumberInMemory(sentinel string, rf *v1alpha1.RedisFailover, auth *util.AuthConfig) error {
	nSlaves, err := r.RedisClient.GetNumberSentinelSlavesInMemory(sentinel, auth)
	if err != nil {
		return err
	} else if nSlaves != rf.Spec.Redis.Replicas-1 {
		return errors.New("sentinel's slaves in memory mismatch")
	}
	return nil
}

func (r RedisFailoverChecker) CheckSentinelMonitor(sentinel string, monitor string, auth *util.AuthConfig) error {
	actualMonitorIP, err := r.RedisClient.GetSentinelMonitor(sentinel, auth)
	if err != nil {
		return err
	}
	if actualMonitorIP != monitor {
		return errors.New("the monitor on the sentinel config does not match with the expected one")
	}
	return nil
}

func (r RedisFailoverChecker) GetMasterIP(rf *v1alpha1.RedisFailover, auth *util.AuthConfig) (string, error) {
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

func (r RedisFailoverChecker) GetNumberMasters(rf *v1alpha1.RedisFailover, auth *util.AuthConfig) (int, error) {
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
	return nMasters, nil
}

func (r RedisFailoverChecker) GetRedisesIPs(rf *v1alpha1.RedisFailover, auth *util.AuthConfig) ([]string, error) {
	redisips := []string{}
	rps, err := r.K8SService.GetStatefulSetPods(rf.Namespace, util.GetRedisName(rf))
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

func (r RedisFailoverChecker) GetSentinelsIPs(rf *v1alpha1.RedisFailover) ([]string, error) {
	sentinels := []string{}
	rps, err := r.K8SService.GetDeploymentPods(rf.Namespace, util.GetSentinelName(rf))
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

func (r RedisFailoverChecker) GetMinimumRedisPodTime(rf *v1alpha1.RedisFailover) (time.Duration, error) {
	minTime := 100000 * time.Hour // More than ten years
	rps, err := r.K8SService.GetStatefulSetPods(rf.Namespace, util.GetRedisName(rf))
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

func (r RedisFailoverChecker) CheckRedisConfig(rf *v1alpha1.RedisFailover, addr string, auth *util.AuthConfig) error {
	panic("implement me")
}
