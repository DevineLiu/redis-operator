package proxyservice

import (
	middlev1alpha1 "github.com/DevineLiu/redis-operator/apis/middle/v1alpha1"
	"github.com/DevineLiu/redis-operator/controllers/middle/client/k8s"
	util2 "github.com/DevineLiu/redis-operator/controllers/util"
	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type RedisProxyClient interface {
	EnsureRedisProxyService(rp *middlev1alpha1.RedisProxy, labels map[string]string, or []metav1.OwnerReference) error
	EnsureRedisProxyDeployment(rp *middlev1alpha1.RedisProxy, labels map[string]string, or []metav1.OwnerReference) error
	EnsureRedisProxyConfigMap(rp *middlev1alpha1.RedisProxy, labels map[string]string, or []metav1.OwnerReference) error
	EnsureRedisProxyNodePortService(rp *middlev1alpha1.RedisProxy, labels map[string]string, or []metav1.OwnerReference) error
	EnsureRedisProxyProbeConfigMap(rp *middlev1alpha1.RedisProxy, labels map[string]string, or []metav1.OwnerReference) error
}

type RedisProxyKubeClient struct{
	K8SService   k8s.Services
	Logger       logr.Logger
	StatusWriter client.StatusWriter
	Record       record.EventRecorder
}

func NewRedisProxyKubeClient(k8SService k8s.Services, log logr.Logger, status client.StatusWriter, record record.EventRecorder) *RedisProxyKubeClient {
	return &RedisProxyKubeClient{K8SService: k8SService, Logger: log, StatusWriter: status, Record: record}
}


func (r RedisProxyKubeClient) EnsureRedisProxyService(rp *middlev1alpha1.RedisProxy, labels map[string]string, ownrf []metav1.OwnerReference) error {
	svc := generateRedisProxyService(rp, labels, ownrf)
	return r.K8SService.CreateIfNotExistsService(rp.Namespace, svc)
}

func (r RedisProxyKubeClient) EnsureRedisProxyDeployment(rp *middlev1alpha1.RedisProxy, labels map[string]string, ownrf []metav1.OwnerReference) error {
	if err := r.ensurePodDisruptionBudget(rp, util2.RedisName, util2.RedisRoleName, labels, ownrf); err != nil {
		return err
	}
	old_deploy, err := r.K8SService.GetDeployment(rp.Namespace, util2.GetRedisProxyName(rp))
	if err != nil {
		if errors.IsNotFound(err) {
			deploy := generateRedisProxyDeployment(rp, labels, ownrf)
			return r.K8SService.CreateDeployment(rp.Namespace, deploy)
		}
		return err
	}
	_ = old_deploy
	return nil
}

func (r RedisProxyKubeClient) EnsureRedisProxyConfigMap(rp *middlev1alpha1.RedisProxy, labels map[string]string, ownrf []metav1.OwnerReference) error {
	password := ""
	if rp.Spec.Auth.SecretPath != "" {
		 secret, err := r.K8SService.GetSecret(rp.Namespace, rp.Spec.Auth.SecretPath)
		 if err != nil {
		 	return err
		 }
		 passwd := secret.Data["password"]
		 password = string(passwd)
	}
	cm := generateRedisProxyConfigMap(rp, labels, ownrf,password)
	return r.K8SService.CreateIfNotExistsConfigMap(rp.Namespace, cm)
}

func (r RedisProxyKubeClient) EnsureRedisProxyNodePortService(rp *middlev1alpha1.RedisProxy, labels map[string]string, ownrf []metav1.OwnerReference) error {
	svc := generateRedisProxyNodePortService(rp, labels, ownrf)
	return r.K8SService.CreateIfNotExistsService(rp.Namespace, svc)
}

func (RedisProxyKubeClient) EnsureRedisProxyProbeConfigMap(rp *middlev1alpha1.RedisProxy, labels map[string]string, or []metav1.OwnerReference) error {
	panic("implement me")
}

func generateSelectorLabels(component, name string) map[string]string {
	return map[string]string{
		"app.kubernetes.io/part-of":   util2.AppLabel,
		"app.kubernetes.io/component": component,
		"app.kubernetes.io/name":      name,
	}
}

func (r RedisProxyKubeClient) ensurePodDisruptionBudget(rp *middlev1alpha1.RedisProxy, name string, component string, labels map[string]string, ownerRefs []metav1.OwnerReference) error {
	name = util2.GenerateName(name, rp.Name)
	namespace := rp.Namespace

	minAvailable := intstr.FromInt(2)
	labels = util2.MergeMap(labels, generateSelectorLabels(component, rp.Name))

	pdb := generatePodDisruptionBudget(name, namespace, labels, ownerRefs, minAvailable)

	return r.K8SService.CreateIfNotExistsPodDisruptionBudget(namespace, pdb)
}