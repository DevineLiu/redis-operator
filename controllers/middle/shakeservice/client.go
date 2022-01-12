package shakeservice

import (
	middlev1alpha1 "github.com/DevineLiu/redis-operator/apis/middle/v1alpha1"
	"github.com/DevineLiu/redis-operator/controllers/middle/client/k8s"
	util2 "github.com/DevineLiu/redis-operator/controllers/util"
	"github.com/go-logr/logr"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type RedisShakeClient interface {
	EnsureRedisShakeService(rs *middlev1alpha1.RedisShake, labels map[string]string, or []metav1.OwnerReference) error
	EnsureRedisShakeDeployment(rs *middlev1alpha1.RedisShake, labels map[string]string, or []metav1.OwnerReference) error
	EnsureRedisShakeConfigMap(rs *middlev1alpha1.RedisShake, labels map[string]string, or []metav1.OwnerReference) error
	EnsureRedisShakeInitConfigMap(rs *middlev1alpha1.RedisShake, labels map[string]string, or []metav1.OwnerReference) error
	//	EnsureRedisShakeProbeConfigMap(rs *middlev1alpha1.RedisShake, labels map[string]string, or []metav1.OwnerReference) error
}

type RedisShakeKubeClient struct {
	K8SService   k8s.Services
	Logger       logr.Logger
	StatusWriter client.StatusWriter
	Record       record.EventRecorder
}

func NewRedisShakeKubeClient(k8SService k8s.Services, log logr.Logger, status client.StatusWriter, record record.EventRecorder) *RedisShakeKubeClient {
	return &RedisShakeKubeClient{K8SService: k8SService, Logger: log, StatusWriter: status, Record: record}
}

func (r RedisShakeKubeClient) EnsureRedisShakeService(rs *middlev1alpha1.RedisShake, labels map[string]string, or []metav1.OwnerReference) error {
	svc := generateRedisShakeService(rs, labels, or)
	return r.K8SService.CreateIfNotExistsService(rs.Namespace, svc)
}

func (r RedisShakeKubeClient) EnsureRedisShakeDeployment(rs *middlev1alpha1.RedisShake, labels map[string]string, or []metav1.OwnerReference) error {
	deploy := generateRedisShakeDeployment(rs, labels, or)
	return r.K8SService.CreateOrUpdateDeployment(rs.Namespace, deploy)
}

func (r RedisShakeKubeClient) EnsureRedisShakeConfigMap(rs *middlev1alpha1.RedisShake, labels map[string]string, or []metav1.OwnerReference) error {
	cm := generateRedisShakeConfigMap(rs, labels, or)
	return r.K8SService.CreateIfNotExistsConfigMap(rs.Namespace, cm)
}

func (r RedisShakeKubeClient) EnsureRedisShakeInitConfigMap(rs *middlev1alpha1.RedisShake, labels map[string]string, or []metav1.OwnerReference) error {
	cm := generateRedisShakeInitConfigMap(rs, labels, or)
	return r.K8SService.CreateIfNotExistsConfigMap(rs.Namespace, cm)
}

func generateSelectorLabels(component, name string) map[string]string {
	return map[string]string{
		"app.kubernetes.io/part-of":   util2.AppLabel,
		"app.kubernetes.io/component": component,
		"app.kubernetes.io/name":      name,
	}
}
