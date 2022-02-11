package k8s

import (
	"context"

	"github.com/go-logr/logr"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type RBAC interface {
	GetRoleBinding(namespace string, name string) (*rbacv1.RoleBinding, error)
	CreateRoleBinding(namespace string, rb *rbacv1.RoleBinding) error
	GetRole(namespace string, name string) (*rbacv1.Role, error)
	CreateRole(namespace string, role *rbacv1.Role) error
}

type RBACOption struct {
	client client.Client
	logger logr.Logger
}

func NewRBAC(kubeClient client.Client, logger logr.Logger) RBAC {
	logger = logger.WithValues("service", "k8s.rbac")
	return &RBACOption{
		client: kubeClient,
		logger: logger,
	}
}

func (s *RBACOption) GetRoleBinding(namespace string, name string) (*rbacv1.RoleBinding, error) {
	rb := &rbacv1.RoleBinding{}
	err := s.client.Get(context.TODO(), types.NamespacedName{
		Name:      name,
		Namespace: namespace,
	}, rb)

	if err != nil {
		return nil, err
	}
	return rb, err
}

func (s *RBACOption) CreateRoleBinding(namespace string, rb *rbacv1.RoleBinding) error {
	err := s.client.Create(context.TODO(), rb)
	if err != nil {
		return err
	}
	s.logger.WithValues("namespace", namespace, "roleBindingName", rb.Name).Info("roleBinding created")
	return nil
}

func (s *RBACOption) GetRole(namespace string, name string) (*rbacv1.Role, error) {
	role := &rbacv1.Role{}
	err := s.client.Get(context.TODO(), types.NamespacedName{
		Name:      name,
		Namespace: namespace,
	}, role)

	if err != nil {
		return nil, err
	}
	return role, err
}

func (s *RBACOption) CreateRole(namespace string, role *rbacv1.Role) error {
	err := s.client.Create(context.TODO(), role)
	if err != nil {
		return err
	}
	s.logger.WithValues("namespace", namespace, "roleName", role.Name).Info("role created")
	return nil
}
