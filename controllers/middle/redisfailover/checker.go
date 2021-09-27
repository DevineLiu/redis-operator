package redisfailover

import (
	"context"
	"errors"
	"fmt"
	middlev1alpha1 "github.com/DevineLiu/redis-operator/apis/middle/v1alpha1"
	"github.com/DevineLiu/redis-operator/util"
	v1 "k8s.io/api/core/v1"
	"time"
)

func (r RedisFailoverHandler) CheckAndHeal(rf *middlev1alpha1.RedisFailover) error {
	if err := r.RfChecker.CheckRedisNumber(rf); err != nil {
		r.Record.Event(rf, v1.EventTypeNormal, "WaitPodReady", "waiting for all redis pods ready")
		r.Logger.WithValues("namespace", rf.Namespace, "name", rf.Name).V(2).Info("waiting all redis instance ready")
		rf.Status.SetWaitingPodReadyCondition("waiting pod ready")
		r.StatusWriter.Update(context.TODO(), rf)
		return err
	}

	if err := r.RfChecker.CheckSentinelNumber(rf); err != nil {
		r.Record.Event(rf, v1.EventTypeWarning, "Error", err.Error())
		return nil
	}

	//TODO auth impl
	auth := util.AuthConfig{}

	nMasters, err := r.RfChecker.GetNumberMasters(rf, &auth)
	if err != nil {
		return err
	}
	switch nMasters {
	case 0:
		redisesIP, err := r.RfChecker.GetRedisesIPs(rf, &auth)
		if err != nil {
			return err
		}
		if len(redisesIP) == 1 {
			if err := r.RfHealer.MakeMaster(redisesIP[0], &auth); err != nil {
				return err
			}
			break
		}
		if _, err := r.RfChecker.GetMinimumRedisPodTime(rf); err != nil {
			return err
		}
		if err := r.RfHealer.SetOldestAsMaster(rf, &auth); err != nil {
			return err
		}
	case 1:
		break
	default:
		return errors.New("more than one master, fix manually")
	}
	master, err := r.RfChecker.GetMasterIP(rf, &auth)
	if err != nil {
		return err
	}
	if err := r.RfChecker.CheckAllSlavesFromMaster(master, rf, &auth); err != nil {
		if err := r.RfHealer.SetMasterOnAll(master, rf, &auth); err != nil {
			return err
		}
	}
	if err = r.setRedisConfig(rf, &auth); err != nil {
		return err
	}
	sentinels, err := r.RfChecker.GetSentinelsIPs(rf)
	if err != nil {
		return err
	}
	for _, sip := range sentinels {
		if err := r.RfChecker.CheckSentinelMonitor(sip, master, &auth); err != nil {
			if err := r.RfHealer.NewSentinelMonitor(sip, master, rf, &auth); err != nil {
				return err
			}
		}
	}
	for _, sip := range sentinels {
		if err := r.RfChecker.CheckSentinelSlavesNumberInMemory(sip, rf, &auth); err != nil {
			if err := r.RfHealer.RestoreSentinel(sip, &auth); err != nil {
				return err
			}
			if err := r.waitRestoreSentinelSlavesOK(sip, rf, &auth); err != nil {
				return err
			}
		}
	}
	for _, sip := range sentinels {
		if err := r.RfChecker.CheckSentinelNumberInMemory(sip, rf, &auth); err != nil {
			if err := r.RfHealer.RestoreSentinel(sip, &auth); err != nil {
				return err
			}
		}
	}
	if err = r.setSentinelConfig(rf, &auth, sentinels); err != nil {
		return err
	}

	return nil
}

func (r *RedisFailoverHandler) setSentinelConfig(rf *middlev1alpha1.RedisFailover, auth *util.AuthConfig, sentinels []string) error {
	for _, sip := range sentinels {
		if err := r.RfHealer.SetSentinelCustomConfig(sip, rf, auth); err != nil {
			return err
		}
	}
	return nil
}

func (r *RedisFailoverHandler) setRedisConfig(rf *middlev1alpha1.RedisFailover, auth *util.AuthConfig) error {
	redises, err := r.RfChecker.GetRedisesIPs(rf, auth)
	if err != nil {
		return err
	}
	for _, rip := range redises {
		if err := r.RfChecker.CheckRedisConfig(rf, rip, auth); err != nil {
			r.Record.Event(rf, v1.EventTypeWarning, "CheckConfigErr", err.Error())
			if err := r.RfHealer.SetRedisCustomConfig(rip, rf, auth); err != nil {
				return err
			}
		}
	}
	return nil
}

func (r *RedisFailoverHandler) waitRestoreSentinelSlavesOK(sentinel string, rf *middlev1alpha1.RedisFailover, auth *util.AuthConfig) error {
	timer := time.NewTimer(30 * time.Second)
	defer timer.Stop()
	for {
		select {
		case <-timer.C:
			return fmt.Errorf("wait for resetore sentinel slave timeout")
		default:
			if err := r.RfChecker.CheckSentinelSlavesNumberInMemory(sentinel, rf, auth); err != nil {
				time.Sleep(5 * time.Second)
			} else {
				return nil
			}
		}
	}
}
