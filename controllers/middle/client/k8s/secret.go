package k8s

import (
	"context"
	"github.com/go-logr/logr"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type Secret interface {
	GetSecret(namespace string,name string) (*v1.Secret,error)
	CreateSecret(namespace string, secret *v1.Secret) error
	UpdateSecret(namespace string, secret *v1.Secret) error
	CreateOrUpdateSecret(namespace string, secret *v1.Secret) error
	DeleteSecret(namespace string,name string) error
	ListSecret(namespace string) (*v1.SecretList, error)
	CreateIfNotExistsSecret(namespace string,secret *v1.Secret) error
}

func (s *SecretOption) GetSecret(namespace string, name string) (*v1.Secret, error) {
	secret := &v1.Secret{}
	err := s.client.Get(context.TODO(), types.NamespacedName{
		Name:      name,
		Namespace: namespace,
	}, secret)
	return secret,err
}

func (s *SecretOption) CreateSecret(namespace string, secret *v1.Secret) error {
	err := s.client.Create(context.TODO(), secret)
	return err
}

func (s *SecretOption) UpdateSecret(namespace string, secret *v1.Secret) error {
	err := s.client.Update(context.TODO(), secret)
	return err
}

func (s *SecretOption) CreateOrUpdateSecret(namespace string, secret *v1.Secret) error {
	secret,err := s.GetSecret(namespace,secret.Name)
	if err != nil {
		if errors.IsNotFound(err) {
			return s.CreateSecret(namespace, secret)
		}
		return err
	}
	secret.ResourceVersion = secret.ResourceVersion
	return s.UpdateSecret(namespace, secret)
}


func (s *SecretOption) DeleteSecret(namespace string, name string) error {
	secret := &v1.Secret{}
	if err := s.client.Get(context.TODO(), types.NamespacedName{
		Name:      name,
		Namespace: namespace,
	}, secret); err != nil {
		return err
	}
	return s.client.Delete(context.TODO(),secret)
}

func (s *SecretOption) ListSecret(namespace string) (*v1.SecretList, error) {
	secret := &v1.SecretList{}
	listOps := &client.ListOptions{
		Namespace: namespace,
	}
	err := s.client.List(context.TODO(), secret, listOps)
	return secret,err
}

func (s *SecretOption) CreateIfNotExistsSecret(namespace string, secret *v1.Secret) error {
	if _, err := s.GetSecret(namespace, secret.Name); err != nil {
		if errors.IsNotFound(err) {
			return s.CreateSecret(namespace, secret)
		}
		return err
	}
	return nil
}

type SecretOption struct {
	client client.Client
	logger logr.Logger
}

func NewSecret(kubeClient client.Client, logger logr.Logger) Secret {
	logger = logger.WithValues("service", "k8s.configMap")
	return &SecretOption{
		client: kubeClient,
		logger: logger,
	}
}