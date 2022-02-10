package util

import (
	"fmt"
	"strings"

	databasesv1 "github.com/DevineLiu/redis-operator/apis/databases/v1"
)

const (
	BaseName               string = "rf"
	SentinelName           string = "s"
	RedisName              string = "r"
	AppLabel               string = "redis-failover"
	RedisRoleName          string = "redis"
	SentinelRoleName       string = "sentinel"
	RedisConfigFileName    string = "redis.conf"
	RedisInitScript        string = "init.sh"
	SentinelConfigFileName string = "sentinel.conf"
	HostnameTopologyKey    string = "kubernetes.io/hostname"
	RedisBackupVolumeName  string = "backup-data"
	RestoreContainerName   string = "restore"
	RestoreDefaultImage    string = "build-harbor.alauda.cn/middleware/redis-backup:v3.7.0"
)

func GenerateName(typeName, metaName string) string {
	return fmt.Sprintf("%s%s-%s", BaseName, typeName, metaName)
}

func GetSentinelName(rf *databasesv1.RedisFailover) string {
	return GenerateName(SentinelName, rf.Name)
}

func GetRedisName(rf *databasesv1.RedisFailover) string {
	return GenerateName(RedisName, rf.Name)
}

func GetSentinelHeadlessSvc(rf *databasesv1.RedisFailover) string {
	return GenerateName(SentinelName, fmt.Sprintf("%s-%s", rf.Name, "hl"))
}

func GetSentinelReadinessConfigmap(rf *databasesv1.RedisFailover) string {
	return GenerateName(fmt.Sprintf("%s-%s", SentinelName, "r"), rf.Name)
}

func GetRedisShutdownConfigMapName(rf *databasesv1.RedisFailover) string {
	return GenerateName(fmt.Sprintf("%s-%s", RedisName, "s"), rf.Name)
}

func GetRedisSecretName(rf *databasesv1.RedisFailover) string {
	return GenerateName(fmt.Sprintf("%s-%s", RedisName, "p"), rf.Name)
}

func GetRedisNodePortSvc(rf *databasesv1.RedisFailover) string {
	return GenerateName(fmt.Sprintf("%s-%s", RedisName, "n"), rf.Name)
}

func GetClaimName(backupDestination string) string {
	names := strings.Split(backupDestination, "/")
	if len(names) != 2 {
		return ""
	}
	return names[1]
}
