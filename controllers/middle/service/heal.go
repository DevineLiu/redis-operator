package service

import (
	"errors"
	middlev1alpha1 "github.com/DevineLiu/redis-operator/apis/middle/v1alpha1"
	"github.com/DevineLiu/redis-operator/controllers/middle/client/k8s"
	"github.com/DevineLiu/redis-operator/controllers/middle/client/redis"
	"github.com/DevineLiu/redis-operator/util"
	"github.com/go-logr/logr"
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sort"
	"strconv"
)

type RedisFailoverHeal interface {
	MakeMaster(ip string, auth *util.AuthConfig) error
	SetOldestAsMaster(rf *middlev1alpha1.RedisFailover, auth *util.AuthConfig) error
	SetMasterOnAll(masterIP string, rf *middlev1alpha1.RedisFailover, auth *util.AuthConfig) error
	NewSentinelMonitor(ip string, monitor string, rf *middlev1alpha1.RedisFailover, auth *util.AuthConfig) error
	RestoreSentinel(ip string, auth *util.AuthConfig) error
	SetSentinelCustomConfig(ip string, rf *middlev1alpha1.RedisFailover, auth *util.AuthConfig) error
	SetRedisCustomConfig(ip string, rf *middlev1alpha1.RedisFailover, auth *util.AuthConfig) error
}

type RedisFailoverHealer struct {
	K8SService   k8s.Services
	Logger       logr.Logger
	StatusWriter client.StatusWriter
	Record       record.EventRecorder
	RedisClient  redis.Client
}

func NewRedisFailoverHealer(k8SService k8s.Services, log logr.Logger, status client.StatusWriter, record record.EventRecorder, rc redis.Client) *RedisFailoverHealer {
	return &RedisFailoverHealer{
		K8SService: k8SService, Logger: log, StatusWriter: status, Record: record, RedisClient: rc,
	}
}

func (r RedisFailoverHealer) MakeMaster(ip string, auth *util.AuthConfig) error {
	return r.RedisClient.MakeMaster(ip, auth)
}

func (r RedisFailoverHealer) SetOldestAsMaster(rf *middlev1alpha1.RedisFailover, auth *util.AuthConfig) error {
	ssp, err := r.K8SService.GetStatefulSetPods(rf.Namespace, util.GetRedisName(rf))
	if err != nil {
		return err
	}
	if len(ssp.Items) < 1 {
		return errors.New("number of redis pods are 0")
	}

	// Order the pods so we start by the oldest one
	sort.Slice(ssp.Items, func(i, j int) bool {
		return ssp.Items[i].CreationTimestamp.Before(&ssp.Items[j].CreationTimestamp)
	})

	newMasterIP := ""
	for _, pod := range ssp.Items {
		if newMasterIP == "" {
			newMasterIP = pod.Status.PodIP
			if err := r.RedisClient.MakeMaster(newMasterIP, auth); err != nil {
				return err
			}
		} else {
			if err := r.RedisClient.MakeSlaveOf(pod.Status.PodIP, newMasterIP, auth); err != nil {
				return err
			}
		}
	}
	return nil
}

func (r RedisFailoverHealer) SetMasterOnAll(masterIP string, rf *middlev1alpha1.RedisFailover, auth *util.AuthConfig) error {
	ssp, err := r.K8SService.GetStatefulSetPods(rf.Namespace, util.GetRedisName(rf))
	if err != nil {
		return err
	}
	for _, pod := range ssp.Items {
		if pod.Status.PodIP == masterIP {
			if err := r.RedisClient.MakeMaster(masterIP, auth); err != nil {
				return err
			}
		} else {
			if err := r.RedisClient.MakeSlaveOf(pod.Status.PodIP, masterIP, auth); err != nil {
				return err
			}
		}
	}
	return nil
}

func (r RedisFailoverHealer) NewSentinelMonitor(ip string, monitor string, rf *middlev1alpha1.RedisFailover, auth *util.AuthConfig) error {
	quorum := strconv.Itoa(int(rf.Spec.Sentinel.Replicas/2 + 1))
	return r.RedisClient.MonitorRedis(ip, monitor, quorum, auth)
}

func (r RedisFailoverHealer) RestoreSentinel(ip string, auth *util.AuthConfig) error {
	return r.RedisClient.ResetSentinel(ip, auth)
}

func (r RedisFailoverHealer) SetSentinelCustomConfig(ip string, rf *middlev1alpha1.RedisFailover, auth *util.AuthConfig) error {
	if len(rf.Spec.Sentinel.CustomConfig) == 0 {
		return nil
	}
	return r.RedisClient.SetCustomSentinelConfig(ip, rf.Spec.Sentinel.CustomConfig, auth)
}

func (r RedisFailoverHealer) SetRedisCustomConfig(ip string, rf *middlev1alpha1.RedisFailover, auth *util.AuthConfig) error {
	if len(rf.Spec.Redis.CustomConfig) == 0 && len(auth.Password) == 0 {
		return nil
	}
	return r.RedisClient.SetCustomRedisConfig(ip, rf.Spec.Redis.CustomConfig, auth)
}
