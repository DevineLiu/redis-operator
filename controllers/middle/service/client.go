package service

import (
	middlev1alpha1 "github.com/DevineLiu/redis-operator/apis/middle/v1alpha1"
	"github.com/DevineLiu/redis-operator/controllers/middle/client/k8s"
	util2 "github.com/DevineLiu/redis-operator/controllers/util"
	"github.com/go-logr/logr"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/tools/record"
	"reflect"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"time"
)

type RedisFailoverClient interface {
	EnsureSentinelService(rf *middlev1alpha1.RedisFailover, labels map[string]string, or []metav1.OwnerReference) error
	EnsureSentinelHeadlessService(rf *middlev1alpha1.RedisFailover, labels map[string]string, ownerRefs []metav1.OwnerReference) error
	EnsureSentinelConfigMap(rf *middlev1alpha1.RedisFailover, labels map[string]string, ownerRefs []metav1.OwnerReference) error
	EnsureSentinelProbeConfigMap(rf *middlev1alpha1.RedisFailover, labels map[string]string, ownerRefs []metav1.OwnerReference) error
	EnsureSentinelDeployment(rf *middlev1alpha1.RedisFailover, labels map[string]string, ownerRefs []metav1.OwnerReference) error
	EnsureRedisStatefulSet(rf *middlev1alpha1.RedisFailover, labels map[string]string, ownerRefs []metav1.OwnerReference) error
	EnsureRedisService(rf *middlev1alpha1.RedisFailover, labels map[string]string, ownerRefs []metav1.OwnerReference) error
	EnsureRedisNodePortService(rf *middlev1alpha1.RedisFailover, labels map[string]string, ownerRefs []metav1.OwnerReference) error
	EnsureRedisShutdownConfigMap(rf *middlev1alpha1.RedisFailover, labels map[string]string, ownerRefs []metav1.OwnerReference) error
	EnsureRedisConfigMap(rf *middlev1alpha1.RedisFailover, labels map[string]string, ownerRefs []metav1.OwnerReference) error
	EnsureNotPresentRedisService(rf *middlev1alpha1.RedisFailover) error
	EnsurePasswordSecrets(rf *middlev1alpha1.RedisFailover, labels map[string]string, ownerRefs []metav1.OwnerReference) error
}

type RedisFailoverKubeClient struct {
	K8SService   k8s.Services
	Logger       logr.Logger
	StatusWriter client.StatusWriter
	Record       record.EventRecorder
}

func generateSelectorLabels(component, name string) map[string]string {
	return map[string]string{
		"app.kubernetes.io/part-of":   util2.AppLabel,
		"app.kubernetes.io/component": component,
		"app.kubernetes.io/name":      name,
	}
}

func NewRedisFailoverKubeClient(k8SService k8s.Services, log logr.Logger, status client.StatusWriter, record record.EventRecorder) *RedisFailoverKubeClient {
	return &RedisFailoverKubeClient{K8SService: k8SService, Logger: log, StatusWriter: status, Record: record}
}

func (r RedisFailoverKubeClient) EnsureSentinelService(rf *middlev1alpha1.RedisFailover, labels map[string]string, ownerRefs []metav1.OwnerReference) error {
	svc := generateSentinelService(rf, labels, ownerRefs)
	return r.K8SService.CreateIfNotExistsService(rf.Namespace, svc)

}

func (r RedisFailoverKubeClient) EnsureSentinelHeadlessService(rf *middlev1alpha1.RedisFailover, labels map[string]string, ownerRefs []metav1.OwnerReference) error {
	svc := newHeadLessSvcForCR(rf, labels, ownerRefs)
	return r.K8SService.CreateIfNotExistsService(rf.Namespace, svc)
}

func (r RedisFailoverKubeClient) EnsureSentinelConfigMap(rf *middlev1alpha1.RedisFailover, labels map[string]string, ownerRefs []metav1.OwnerReference) error {
	cm := generateSentinelConfigMap(rf, labels, ownerRefs)
	return r.K8SService.CreateIfNotExistsConfigMap(rf.Namespace, cm)
}

func (r RedisFailoverKubeClient) EnsureSentinelProbeConfigMap(rf *middlev1alpha1.RedisFailover, labels map[string]string, ownerRefs []metav1.OwnerReference) error {
	cm := generateSentinelReadinessProbeConfigMap(rf, labels, ownerRefs)
	return r.K8SService.CreateIfNotExistsConfigMap(rf.Namespace, cm)
}

func (r RedisFailoverKubeClient) EnsureSentinelDeployment(rf *middlev1alpha1.RedisFailover, labels map[string]string, ownerRefs []metav1.OwnerReference) error {
	if err := r.ensurePodDisruptionBudget(rf, util2.RedisName, util2.RedisRoleName, labels, ownerRefs); err != nil {
		return err
	}
	oldSs, err := r.K8SService.GetDeployment(rf.Namespace, util2.GetSentinelName(rf))
	if err != nil {
		if errors.IsNotFound(err) {
			deploy := generateSentinelDeployment(rf, labels, ownerRefs)
			return r.K8SService.CreateDeployment(rf.Namespace, deploy)
		}
		return err
	}
	if shouldUpdateRedis(rf.Spec.Sentinel.Resources, oldSs.Spec.Template.Spec.Containers[0].Resources,
		rf.Spec.Sentinel.Replicas, *oldSs.Spec.Replicas) {
		deploy := generateSentinelDeployment(rf, labels, ownerRefs)
		return r.K8SService.UpdateDeployment(rf.Namespace, deploy)
	}
	return nil
}

func (r RedisFailoverKubeClient) EnsureRedisStatefulSet(rf *middlev1alpha1.RedisFailover, labels map[string]string, ownerRefs []metav1.OwnerReference) error {
	if err := r.ensurePodDisruptionBudget(rf, util2.SentinelName, util2.SentinelRoleName, labels, ownerRefs); err != nil {
		return err
	}
	oldSs, err := r.K8SService.GetStatefulSet(rf.Namespace, util2.GetRedisName(rf))
	if err != nil {
		// If no resource we need to create.
		if errors.IsNotFound(err) {
			ss := generateRedisStatefulSet(rf, labels, ownerRefs)
			return r.K8SService.CreateStatefulSet(rf.Namespace, ss)
		}

		return err
	}
	if shouldUpdateRedis(rf.Spec.Redis.Resources, oldSs.Spec.Template.Spec.Containers[0].Resources,
		rf.Spec.Redis.Replicas, *oldSs.Spec.Replicas) || exporterChanged(rf, oldSs) {
		ss := generateRedisStatefulSet(rf, labels, ownerRefs)
		return r.K8SService.UpdateStatefulSet(rf.Namespace, ss)
	}
	return nil
}

func (r RedisFailoverKubeClient) EnsureRedisService(rf *middlev1alpha1.RedisFailover, labels map[string]string, ownerRefs []metav1.OwnerReference) error {
	svc := generateRedisService(rf, labels, ownerRefs)
	return r.K8SService.CreateIfNotExistsService(rf.Namespace, svc)
}

func (r RedisFailoverKubeClient) EnsureRedisNodePortService(rf *middlev1alpha1.RedisFailover, labels map[string]string, ownerRefs []metav1.OwnerReference) error {
	panic("implement me")
}

func (r RedisFailoverKubeClient) EnsureRedisShutdownConfigMap(rf *middlev1alpha1.RedisFailover, labels map[string]string, ownerRefs []metav1.OwnerReference) error {
	if rf.Spec.Redis.ShutdownConfigMap != "" {
		if _, err := r.K8SService.GetConfigMap(rf.Namespace, rf.Spec.Redis.ShutdownConfigMap); err != nil {
			return err
		}
	} else {
		cm := generateRedisShutdownConfigMap(rf, labels, ownerRefs)
		return r.K8SService.CreateIfNotExistsConfigMap(rf.Namespace, cm)
	}
	return nil
}

func (r RedisFailoverKubeClient) EnsureRedisConfigMap(rf *middlev1alpha1.RedisFailover, labels map[string]string, ownerRefs []metav1.OwnerReference) error {
	panic("implement me")
}

func (r RedisFailoverKubeClient) EnsureNotPresentRedisService(rf *middlev1alpha1.RedisFailover) error {
	panic("implement me")
}

func (r RedisFailoverKubeClient) EnsurePasswordSecrets(rf *middlev1alpha1.RedisFailover, labels map[string]string, ownerRefs []metav1.OwnerReference) error {
	secret, err := r.K8SService.GetSecret(rf.Namespace, rf.Spec.Auth.SecretPath)
	if err != nil {
		return err
	}
	passwd := secret.Data["password"]
	if len(passwd) <= 0 {
		return nil
	}
	now := time.Now()
	if secretWithVersion, err := r.K8SService.GetSecret(rf.Namespace, util2.GetRedisSecretName(rf)); err != nil {

		if errors.IsNotFound(err) {
			secretWithVersion = &corev1.Secret{}
			secretWithVersion.Name=util2.GetRedisSecretName(rf)
			timestr := now.Format(time.RFC3339)
			secretWithVersion.Data[timestr] = passwd
			secretWithVersion.Labels=labels
			r.K8SService.CreateSecret(rf.Namespace, secretWithVersion)
		} else {
			return err
		}
	} else {
		lastData := []byte{}
		initTime := time.Time{}
		for k, v := range secretWithVersion.Data {
			if time, err := time.Parse(time.RFC3339, k); err == nil {
				if time.After(now) {
					return errors.NewResourceExpired("now timestamp is large than secret's  timestamp")
				}
				if time.After(initTime) {
					initTime = time
					lastData = v
				}
			} else {
				return err
			}

		}
		if !reflect.DeepEqual(lastData, passwd) {
			timestr := now.Format(time.RFC3339)
			secretWithVersion.Data[timestr] = passwd
			r.K8SService.UpdateSecret(rf.Namespace, secretWithVersion)
		}
	}
	return err
}

func (r RedisFailoverKubeClient) ensurePodDisruptionBudget(rf *middlev1alpha1.RedisFailover, name string, component string, labels map[string]string, ownerRefs []metav1.OwnerReference) error {
	name = util2.GenerateName(name, rf.Name)
	namespace := rf.Namespace

	minAvailable := intstr.FromInt(2)
	labels = util2.MergeMap(labels, generateSelectorLabels(component, rf.Name))

	pdb := generatePodDisruptionBudget(name, namespace, labels, ownerRefs, minAvailable)

	return r.K8SService.CreateIfNotExistsPodDisruptionBudget(namespace, pdb)
}

func shouldUpdateRedis(expectResource, containterResource corev1.ResourceRequirements, expectSize, replicas int32) bool {
	if expectSize != replicas {
		return true
	}
	if result := containterResource.Requests.Cpu().Cmp(*expectResource.Requests.Cpu()); result != 0 {
		return true
	}
	if result := containterResource.Requests.Memory().Cmp(*expectResource.Requests.Memory()); result != 0 {
		return true
	}
	if result := containterResource.Limits.Cpu().Cmp(*expectResource.Limits.Cpu()); result != 0 {
		return true
	}
	if result := containterResource.Limits.Memory().Cmp(*expectResource.Limits.Memory()); result != 0 {
		return true
	}
	return false
}

func exporterChanged(rf *middlev1alpha1.RedisFailover, sts *appsv1.StatefulSet) bool {
	if rf.Spec.Redis.Exporter.Enabled {
		for _, container := range sts.Spec.Template.Spec.Containers {
			if container.Name == exporterContainerName {
				return false
			}
		}
		return true
	} else {
		for _, container := range sts.Spec.Template.Spec.Containers {
			if container.Name == exporterContainerName {
				return true
			}
		}
		return false
	}
}
