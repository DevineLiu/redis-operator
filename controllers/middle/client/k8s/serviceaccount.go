package k8s

import (
	"context"

	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type ServiceAccount interface {
	GetServiceAccount(namespace string, name string) (*corev1.ServiceAccount, error)
	CreateServiceAccount(namespace string, sa *corev1.ServiceAccount) error
}

type ServiceAccountOption struct {
	client client.Client
	logger logr.Logger
}

func NewServiceAccount(kubeClient client.Client, logger logr.Logger) ServiceAccount {
	logger = logger.WithValues("service", "k8s.serviceAccount")
	return &ServiceAccountOption{
		client: kubeClient,
		logger: logger,
	}
}

func (s *ServiceAccountOption) GetServiceAccount(namespace string, name string) (*corev1.ServiceAccount, error) {
	sa := &corev1.ServiceAccount{}
	err := s.client.Get(context.TODO(), types.NamespacedName{
		Name:      name,
		Namespace: namespace,
	}, sa)

	if err != nil {
		return nil, err
	}
	return sa, err
}

func (s *ServiceAccountOption) CreateServiceAccount(namespace string, sa *corev1.ServiceAccount) error {
	err := s.client.Create(context.TODO(), sa)
	if err != nil {
		return err
	}
	s.logger.WithValues("namespace", namespace, "serviceAccountName", sa.Name).Info("serviceAccount created")
	return nil
}
