package service

import (
	"reflect"
	"strconv"
	"time"

	databasesv1 "github.com/DevineLiu/redis-operator/apis/databases/v1"
	util2 "github.com/DevineLiu/redis-operator/controllers/databases/util"
	"github.com/DevineLiu/redis-operator/controllers/middle/client/k8s"
	util "github.com/DevineLiu/redis-operator/controllers/util"
	redisbackup "github.com/DevineLiu/redis-operator/extend/redisbackup/v1"
	"github.com/go-logr/logr"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type RedisFailoverClient interface {
	EnsureSentinelService(rf *databasesv1.RedisFailover, labels map[string]string, or []metav1.OwnerReference) error
	EnsureSentinelHeadlessService(rf *databasesv1.RedisFailover, labels map[string]string, ownerRefs []metav1.OwnerReference) error
	EnsureSentinelConfigMap(rf *databasesv1.RedisFailover, labels map[string]string, ownerRefs []metav1.OwnerReference) error
	EnsureSentinelProbeConfigMap(rf *databasesv1.RedisFailover, labels map[string]string, ownerRefs []metav1.OwnerReference) error
	EnsureSentinelDeployment(rf *databasesv1.RedisFailover, labels map[string]string, ownerRefs []metav1.OwnerReference) error
	EnsureRedisStatefulSet(rf *databasesv1.RedisFailover, labels map[string]string, ownerRefs []metav1.OwnerReference) error
	EnsureRedisService(rf *databasesv1.RedisFailover, labels map[string]string, ownerRefs []metav1.OwnerReference) error
	EnsureRedisNodePortService(rf *databasesv1.RedisFailover, labels map[string]string, ownerRefs []metav1.OwnerReference) error
	EnsureRedisShutdownConfigMap(rf *databasesv1.RedisFailover, labels map[string]string, ownerRefs []metav1.OwnerReference) error
	EnsureRedisConfigMap(rf *databasesv1.RedisFailover, labels map[string]string, ownerRefs []metav1.OwnerReference) error
	EnsureNotPresentRedisService(rf *databasesv1.RedisFailover) error
	EnsurePasswordSecrets(rf *databasesv1.RedisFailover, labels map[string]string, ownerRefs []metav1.OwnerReference) error
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

func (r RedisFailoverKubeClient) EnsureSentinelService(rf *databasesv1.RedisFailover, labels map[string]string, ownerRefs []metav1.OwnerReference) error {
	svc := generateSentinelService(rf, labels, ownerRefs)
	return r.K8SService.CreateIfNotExistsService(rf.Namespace, svc)

}

func (r RedisFailoverKubeClient) EnsureSentinelHeadlessService(rf *databasesv1.RedisFailover, labels map[string]string, ownerRefs []metav1.OwnerReference) error {
	svc := newHeadLessSvcForCR(rf, labels, ownerRefs)
	return r.K8SService.CreateIfNotExistsService(rf.Namespace, svc)
}

func (r RedisFailoverKubeClient) EnsureSentinelConfigMap(rf *databasesv1.RedisFailover, labels map[string]string, ownerRefs []metav1.OwnerReference) error {
	cm := generateSentinelConfigMap(rf, labels, ownerRefs)
	return r.K8SService.CreateIfNotExistsConfigMap(rf.Namespace, cm)
}

func (r RedisFailoverKubeClient) EnsureSentinelProbeConfigMap(rf *databasesv1.RedisFailover, labels map[string]string, ownerRefs []metav1.OwnerReference) error {
	cm := generateSentinelReadinessProbeConfigMap(rf, labels, ownerRefs)
	return r.K8SService.CreateIfNotExistsConfigMap(rf.Namespace, cm)
}

func (r RedisFailoverKubeClient) EnsureSentinelDeployment(rf *databasesv1.RedisFailover, labels map[string]string, ownerRefs []metav1.OwnerReference) error {
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

func (r RedisFailoverKubeClient) EnsureRedisStatefulSet(rf *databasesv1.RedisFailover, labels map[string]string, ownerRefs []metav1.OwnerReference) error {
	if err := r.ensurePodDisruptionBudget(rf, util2.SentinelName, util2.SentinelRoleName, labels, ownerRefs); err != nil {
		return err
	}

	var backup *redisbackup.RedisBackup
	if rf.Spec.Redis.Restore.BackupName != "" {
		var err error
		backup, err = r.K8SService.GetRedisBackup(rf.Namespace, rf.Spec.Redis.Restore.BackupName)
		if err != nil {
			return err
		}
	}

	// oldSs, err := r.K8SService.GetStatefulSet(rf.Namespace, util2.GetRedisName(rf))
	// if err != nil {
	// 	// If no resource we need to create.
	// 	if errors.IsNotFound(err) {

	// 		ss := generateRedisStatefulSet(rf, labels, ownerRefs, backup)
	// 		return r.K8SService.CreateStatefulSet(rf.Namespace, ss)
	// 	}

	// 	return err
	// }

	// if shouldUpdateRedis(rf.Spec.Redis.Resources, oldSs.Spec.Template.Spec.Containers[0].Resources,
	// 	rf.Spec.Redis.Replicas, *oldSs.Spec.Replicas) || exporterChanged(rf, oldSs) {
	// 	ss := generateRedisStatefulSet(rf, labels, ownerRefs, backup)
	// 	return r.K8SService.UpdateStatefulSet(rf.Namespace, ss)
	// }

	ss := generateRedisStatefulSet(rf, labels, ownerRefs, backup)
	return r.K8SService.CreateOrUpdateStatefulSet(rf.Namespace, ss)
}

func (r RedisFailoverKubeClient) EnsureRedisService(rf *databasesv1.RedisFailover, labels map[string]string, ownerRefs []metav1.OwnerReference) error {
	svc := generateRedisService(rf, labels, ownerRefs)
	return r.K8SService.CreateIfNotExistsService(rf.Namespace, svc)
}

func (r RedisFailoverKubeClient) EnsureRedisNodePortService(rf *databasesv1.RedisFailover, labels map[string]string, ownerRefs []metav1.OwnerReference) error {
	panic("implement me")
}

func (r RedisFailoverKubeClient) EnsureRedisShutdownConfigMap(rf *databasesv1.RedisFailover, labels map[string]string, ownerRefs []metav1.OwnerReference) error {
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

func (r RedisFailoverKubeClient) EnsureRedisConfigMap(rf *databasesv1.RedisFailover, labels map[string]string, ownerRefs []metav1.OwnerReference) error {
	configmap := generateRedisConfigMap(rf, labels, ownerRefs)
	err := r.K8SService.CreateOrUpdateConfigMap(rf.Namespace, configmap)
	return err
}

func (r RedisFailoverKubeClient) EnsureNotPresentRedisService(rf *databasesv1.RedisFailover) error {
	panic("implement me")
}

func (r RedisFailoverKubeClient) EnsurePasswordSecrets(rf *databasesv1.RedisFailover, labels map[string]string, ownerRefs []metav1.OwnerReference) error {
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
			secretWithVersion.Namespace = rf.Namespace
			secretWithVersion.Name = util2.GetRedisSecretName(rf)
			timestr := now.Unix()
			secretWithVersion.Data = make(map[string][]byte)
			secretWithVersion.Data[strconv.Itoa(int(timestr))] = passwd
			secretWithVersion.Labels = labels
			err := r.K8SService.CreateSecret(rf.Namespace, secretWithVersion)
			if err != nil {
				return err
			}
		} else {
			return err
		}
	} else {
		lastData := []byte{}
		initTime := time.Time{}
		for k, v := range secretWithVersion.Data {
			if time_i, err := strconv.Atoi(k); err == nil {
				time_u := time.Unix(int64(time_i), 0)
				if time_u.After(now) {
					return errors.NewResourceExpired("now timestamp is large than secret's  timestamp")
				}
				if time_u.After(initTime) {
					initTime = time_u
					lastData = v
				}
			} else {
				return err
			}

		}
		if !reflect.DeepEqual(lastData, passwd) {
			timestr := strconv.Itoa(int(now.Unix()))
			secretWithVersion.Data[timestr] = passwd
			r.K8SService.UpdateSecret(rf.Namespace, secretWithVersion)
		}
	}
	return err
}

func (r RedisFailoverKubeClient) ensurePodDisruptionBudget(rf *databasesv1.RedisFailover, name string, component string, labels map[string]string, ownerRefs []metav1.OwnerReference) error {
	name = util2.GenerateName(name, rf.Name)
	namespace := rf.Namespace

	minAvailable := intstr.FromInt(2)
	labels = util.MergeMap(labels, generateSelectorLabels(component, rf.Name))

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

func exporterChanged(rf *databasesv1.RedisFailover, sts *appsv1.StatefulSet) bool {
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
