package v1alpha1

import (
	"errors"
	"fmt"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

const (
	maxNameLength = 48

	defaultRedisNumber     = 3
	defaultSentinelNumber  = 3
	defaultRedisImage      = "redis:5.0.4-alpine"
	defaultRedisProxyImage = "build-harbor.alauda.cn/middleware/redis-proxy:v3.7.0"
	// TODO : set default Slave
	defaultSlavePriority = "1"
)

func (r *RedisFailover) Validate() error {
	if len(r.Name) > maxNameLength {
		return fmt.Errorf("name length can't be higher than %d", maxNameLength)
	}

	if r.Spec.Redis.Replicas == 0 {
		r.Spec.Redis.Replicas = defaultRedisNumber
	} else if r.Spec.Redis.Replicas < defaultRedisNumber {
		return errors.New("number of redis in spec is less than the minimum")
	}

	if r.Spec.Sentinel.Replicas == 0 {
		r.Spec.Sentinel.Replicas = defaultSentinelNumber
	} else if r.Spec.Sentinel.Replicas < defaultSentinelNumber {
		return errors.New("number of sentinel in spec is less than the minimum")
	}

	if r.Spec.Redis.Image == "" {
		r.Spec.Redis.Image = defaultRedisImage
	}
	if r.Spec.Sentinel.Image == "" {
		r.Spec.Sentinel.Image = defaultRedisImage
	}

	if r.Spec.Sentinel.Resources.Size() == 0 {
		r.Spec.Sentinel.Resources = defaultSentinelResource()
	}

	// if r.Spec.Redis.ConfigConfigMap=="" {
	// 	r.Spec.Redis.ConfigConfigMap = make(map[string]string)
	// }
	// if !r.Spec.DisablePersistence {
	// 	enablePersistence(r.Spec.Config)
	// } else {
	// 	disablePersistence(r.Spec.Config)
	// }

	return nil
}

func defaultSentinelResource() v1.ResourceRequirements {
	return v1.ResourceRequirements{
		Requests: v1.ResourceList{
			v1.ResourceCPU:    resource.MustParse("20m"),
			v1.ResourceMemory: resource.MustParse("16Mi"),
		},
		Limits: v1.ResourceList{
			v1.ResourceCPU:    resource.MustParse("100m"),
			v1.ResourceMemory: resource.MustParse("60Mi"),
		},
	}
}

func (rp *RedisProxy) Validate() error {
	if rp.Spec.Replicas <= 0 {
		rp.Spec.Replicas = 1
	}
	if rp.Spec.Resources.Size() == 0 {
		rp.Spec.Resources = defaultSentinelResource()
	}
	if rp.Spec.Image == "" {
		rp.Spec.Image = defaultRedisProxyImage
	}

	if rp.Spec.ProxyInfo.WorkerThreads < 1 {
		rp.Spec.ProxyInfo.WorkerThreads = 4
	}
	if rp.Spec.ProxyInfo.ClientTimeout == 0 {
		rp.Spec.ProxyInfo.ClientTimeout = 120
	}
	if rp.Spec.ProxyInfo.Architecture == "" {
		rp.Spec.ProxyInfo.Architecture = "cluster"
	}
	return nil
}

func defaultProxyResource() v1.ResourceRequirements {
	return v1.ResourceRequirements{
		Requests: v1.ResourceList{
			v1.ResourceCPU:    resource.MustParse("500m"),
			v1.ResourceMemory: resource.MustParse("500Mi"),
		},
		Limits: v1.ResourceList{
			v1.ResourceCPU:    resource.MustParse("1000m"),
			v1.ResourceMemory: resource.MustParse("1024Mi"),
		},
	}
}
