package redisproxy

import (
	middlev1alpha1 "github.com/DevineLiu/redis-operator/apis/middle/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (r *RedisProxyHandler) Ensure(rp *middlev1alpha1.RedisProxy, labels map[string]string, own []metav1.OwnerReference) error {
	if err := r.RpServices.EnsureRedisProxyService(rp, labels, own); err != nil {
		return err
	}
	if err := r.RpServices.EnsureRedisProxyNodePortService(rp, labels, own); err != nil {
		return err
	}
	if err := r.RpServices.EnsureRedisProxyConfigMap(rp, labels, own); err != nil {
		return err
	}
	//if err := r.RpServices.EnsureRedisProxyProbeConfigMap(rp, labels, own); err != nil {
	//	return err
	//}
	if err := r.RpServices.EnsureRedisProxyDeployment(rp, labels, own); err != nil {
		return err
	}
	return nil
}
