package util

import (
	"fmt"
	"github.com/DevineLiu/redis-operator/apis/middle/v1alpha1"
)

const (
	BaseName               = "redis"
	SentinelName           = "-sentinel"
	ProxyName  = "proxy"
	ProxyNodeportName  = "proxyn"
	SentinelRoleName       = "sentinel"
	ProxyRoleName              = "redis-proxy"
	SentinelConfigFileName = "sentinel.conf"
	ProxyConfigFileName = "proxy.conf"
	RedisConfigFileName    = "redis.conf"
	RedisName              = "-redisdb"
	RedisShutdownName      = "r-s"
	RedisRoleName          = "redis"
	AppLabel               = "redis-failover"
	HostnameTopologyKey    = "kubernetes.io/hostname"
)

func GenerateName(typeName, metaName string) string {
	return fmt.Sprintf("%s%s-%s", BaseName, typeName, metaName)
}

func GenerateProxyName(typeName, metaName string) string {
	return fmt.Sprintf("%s-%s",metaName, typeName )
}

func GetRedisName(rf *v1alpha1.RedisFailover) string {
	return GenerateName(RedisName, rf.Name)
}

func GetSentinelName(rf *v1alpha1.RedisFailover) string {
	return GenerateName(SentinelName, rf.Name)
}

func GetRedisProxyName(rp *v1alpha1.RedisProxy) string {
	return  GenerateProxyName(ProxyName,rp.Name)
}

func GetRedisProxyNodePortName(rp *v1alpha1.RedisProxy) string {
	return  GenerateProxyName(ProxyNodeportName,rp.Name)
}


func GetRedisShutdownName(rf *v1alpha1.RedisFailover) string {
	return GenerateName(RedisShutdownName, rf.Name)
}

func GetSentinelHeadlessSvc(rf *v1alpha1.RedisFailover) string {
	return GenerateName("-sentinel-headless", rf.Name)
}

func GetRedisNodePortSvc(rf *v1alpha1.RedisFailover) string {
	return GenerateName("-redis-node-port", rf.Name)
}

func GetRedisSecretName(rf *v1alpha1.RedisFailover) string {
	return GenerateName("-passwd-readonly", rf.Name)
}

func GetSentinelReadinessConfigmap(rf *v1alpha1.RedisFailover) string {
	return GenerateName("-sentinel-readiness", rf.Name)
}

func GetRedisShutdownConfigMapName(rf *v1alpha1.RedisFailover) string {
	if rf.Spec.Redis.ShutdownConfigMap != "" {
		return rf.Spec.Redis.ShutdownConfigMap
	}
	return GetRedisShutdownName(rf)
}


