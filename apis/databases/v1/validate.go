package v1

import (
	"fmt"
)

const (
	maxNameLength                = 48
	defaultRedisNumber           = 2
	defaultSentinelNumber        = 3
	defaultSentinelExporterImage = "leominov/redis_sentinel_exporter:1.3.0"
	defaultExporterImage         = "oliver006/redis_exporter:v1.3.5-alpine"
	defaultImage                 = "build-harbor.alauda.cn/3rdparty/redis:5.0-alpine"
	redis6Image                  = "build-harbor.alauda.cn/3rdparty/redis:6.0-alpine"
)

// Validate set the values by default if not defined and checks if the values given are valid
func (r *RedisFailover) Validate() error {
	if len(r.Name) > maxNameLength {
		return fmt.Errorf("name length can't be higher than %d", maxNameLength)
	}

	if r.Spec.Redis.Image == "" {
		r.Spec.Redis.Image = defaultImage
	}

	if r.Spec.Sentinel.Image == "" {
		r.Spec.Sentinel.Image = defaultImage
	}

	if r.Spec.Redis.Replicas <= 0 {
		r.Spec.Redis.Replicas = defaultRedisNumber
	}

	if r.Spec.Sentinel.Replicas <= 0 {
		r.Spec.Sentinel.Replicas = defaultSentinelNumber
	}

	if r.Spec.Redis.Exporter.Image == "" {
		r.Spec.Redis.Exporter.Image = defaultExporterImage
	}

	if r.Spec.Sentinel.Exporter.Image == "" {
		r.Spec.Sentinel.Exporter.Image = defaultSentinelExporterImage
	}

	// tls only support redis >= 6.0
	if r.Spec.EnableTLS && r.Spec.Redis.Image != redis6Image {
		r.Spec.EnableTLS = false
	}

	return nil
}
