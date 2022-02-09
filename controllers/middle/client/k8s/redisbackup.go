package k8s

import (
	"context"

	"k8s.io/apimachinery/pkg/types"

	redisbackup "github.com/DevineLiu/redis-operator/extend/redisbackup/v1"
	"github.com/go-logr/logr"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// Service the client that knows how to interact with kubernetes to manage them
type RedisBackup interface {
	// GetService get service from kubernetes with namespace and name
	GetRedisBackup(namespace string, name string) (*redisbackup.RedisBackup, error)
}

// ServiceOption is the service client implementation using API calls to kubernetes.
type RedisBackupOption struct {
	client client.Client
	logger logr.Logger
}

// NewService returns a new Service client.
func NewRedisBackup(kubeClient client.Client, logger logr.Logger) RedisBackup {
	logger = logger.WithValues("service", "k8s.service")
	return &RedisBackupOption{
		client: kubeClient,
		logger: logger,
	}
}

// GetService implement the Service.Interface
func (s *RedisBackupOption) GetRedisBackup(namespace string, name string) (*redisbackup.RedisBackup, error) {
	redis_backup := &redisbackup.RedisBackup{}
	err := s.client.Get(context.TODO(), types.NamespacedName{
		Name:      name,
		Namespace: namespace,
	}, redis_backup)

	if err != nil {
		return nil, err
	}
	return redis_backup, err
}
