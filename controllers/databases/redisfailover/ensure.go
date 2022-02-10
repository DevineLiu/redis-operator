package redisfailover

import (
	databasesv1 "github.com/DevineLiu/redis-operator/apis/databases/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (r *RedisFailoverHandler) Ensure(rf *databasesv1.RedisFailover, labels map[string]string, own []metav1.OwnerReference) error {
	if err := r.RfServices.EnsureRedisService(rf, labels, own); err != nil {
		return err
	}
	if err := r.RfServices.EnsureSentinelService(rf, labels, own); err != nil {
		return err
	}
	if err := r.RfServices.EnsureSentinelHeadlessService(rf, labels, own); err != nil {
		return err
	}
	if err := r.RfServices.EnsureSentinelConfigMap(rf, labels, own); err != nil {
		return err
	}
	if err := r.RfServices.EnsureSentinelProbeConfigMap(rf, labels, own); err != nil {
		return err
	}

	if err := r.RfServices.EnsureRedisConfigMap(rf, labels, own); err != nil {
		return err
	}
	if err := r.RfServices.EnsureRedisShutdownConfigMap(rf, labels, own); err != nil {
		return err
	}

	if rf.Spec.Auth.SecretPath != "" {
		if err := r.RfServices.EnsurePasswordSecrets(rf, labels, own); err != nil {
			return err
		}
	}

	if err := r.RfServices.EnsureSentinelDeployment(rf, labels, own); err != nil {
		return err
	}
	if err := r.RfServices.EnsureRedisStatefulSet(rf, labels, own); err != nil {
		return err
	}
	return nil
}
