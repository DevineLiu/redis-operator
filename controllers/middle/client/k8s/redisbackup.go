package k8s

import (
	"context"

	"k8s.io/apimachinery/pkg/types"

	redisbackup "github.com/DevineLiu/redis-operator/extend/redisbackup/v1"
	"github.com/go-logr/logr"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type RedisBackup interface {
	GetRedisBackup(namespace string, name string) (*redisbackup.RedisBackup, error)
	ListRedisBackups(namespace string, listOps client.ListOptions) (*redisbackup.RedisBackupList, error)
	DeleteRedisBackup(namespace string, name string) error
}

type RedisBackupOption struct {
	client client.Client
	logger logr.Logger
}

func NewRedisBackup(kubeClient client.Client, logger logr.Logger) RedisBackup {
	logger = logger.WithValues("service", "k8s.RedisBackup")
	return &RedisBackupOption{
		client: kubeClient,
		logger: logger,
	}
}

func (r *RedisBackupOption) GetRedisBackup(namespace string, name string) (*redisbackup.RedisBackup, error) {
	redis_backup := &redisbackup.RedisBackup{}
	err := r.client.Get(context.TODO(), types.NamespacedName{
		Name:      name,
		Namespace: namespace,
	}, redis_backup)

	if err != nil {
		return nil, err
	}
	return redis_backup, err
}

func (r *RedisBackupOption) ListRedisBackups(namespace string, listOps client.ListOptions) (*redisbackup.RedisBackupList, error) {
	rl := &redisbackup.RedisBackupList{}
	err := r.client.List(context.TODO(), rl, &listOps)
	if err != nil {
		return nil, err
	}
	return rl, err

}

func (r *RedisBackupOption) DeleteRedisBackup(namespace string, name string) error {
	redis_backup := &redisbackup.RedisBackup{}
	if err := r.client.Get(context.TODO(), types.NamespacedName{
		Name:      name,
		Namespace: namespace,
	}, redis_backup); err != nil {
		return err
	}
	return r.client.Delete(context.TODO(), redis_backup)
}
