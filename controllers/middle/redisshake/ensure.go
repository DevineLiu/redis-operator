package redisshake

import (
	middlev1alpha1 "github.com/DevineLiu/redis-operator/apis/middle/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (r *RedisShakeHandler) Ensure(rs *middlev1alpha1.RedisShake, labels map[string]string, own []metav1.OwnerReference) error {
	if err := r.RsServices.EnsureRedisShakeService(rs, labels, own); err != nil {
		return err
	}

	if err := r.RsServices.EnsureRedisShakeConfigMap(rs, labels, own); err != nil {
		return err
	}

	if err := r.RsServices.EnsureRedisShakeInitConfigMap(rs, labels, own); err != nil {
		return err
	}

	if err := r.RsServices.EnsureRedisShakeDeployment(rs, labels, own); err != nil {
		return err
	}

	return nil
}
