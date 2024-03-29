package service

import (
	"fmt"
	"github.com/DevineLiu/redis-operator/apis/middle/v1alpha1"
	util2 "github.com/DevineLiu/redis-operator/controllers/util"
	v1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	policyv1beta1 "k8s.io/api/policy/v1beta1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"strings"
)

const (
	redisShutdownConfigurationVolumeName = "redis-shutdown-config"
	redisStorageVolumeName               = "redis-data"
	exporterContainerName                = "redis-exporter"
	graceTime                            = 30
	redisPasswordEnv                     = "REDIS_PASSWORD"
)

func generateRedisService(rf *v1alpha1.RedisFailover, labels map[string]string, ownerRefs []metav1.OwnerReference) *corev1.Service {
	name := util2.GetRedisName(rf)
	namespace := rf.Namespace

	labels = util2.MergeMap(labels, generateSelectorLabels(util2.RedisRoleName, rf.Name))
	redisTargetPort := intstr.FromInt(6379)
	return &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:            name,
			Namespace:       namespace,
			Labels:          labels,
			OwnerReferences: ownerRefs,
		},
		Spec: corev1.ServiceSpec{
			Type:      corev1.ServiceTypeClusterIP,
			ClusterIP: corev1.ClusterIPNone,
			Ports: []corev1.ServicePort{
				{
					Port:       6379,
					Protocol:   corev1.ProtocolTCP,
					Name:       "redis",
					TargetPort: redisTargetPort,
				},
			},
			Selector: labels,
		},
	}
}

func generateRedisNodePortService(rf *v1alpha1.RedisFailover, labels map[string]string, ownerRefs []metav1.OwnerReference) *corev1.Service {
	namespace := rf.Namespace

	labels = util2.MergeMap(labels, generateSelectorLabels(util2.RedisRoleName, rf.Name))
	redisTargetPort := intstr.FromInt(6379)
	return &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:            util2.GetRedisNodePortSvc(rf),
			Namespace:       namespace,
			Labels:          labels,
			OwnerReferences: ownerRefs,
		},
		Spec: corev1.ServiceSpec{
			Type:      corev1.ServiceTypeNodePort,
			ClusterIP: corev1.ClusterIPNone,
			Ports: []corev1.ServicePort{
				{
					Port:       6379,
					Protocol:   corev1.ProtocolTCP,
					Name:       "redis",
					TargetPort: redisTargetPort,
				},
			},
			Selector: labels,
		},
	}
}

func newHeadLessSvcForCR(rf *v1alpha1.RedisFailover, labels map[string]string, ownerRefs []metav1.OwnerReference) *corev1.Service {
	sentinelPort := corev1.ServicePort{Name: "sentinel", Port: 26379}
	labels = util2.MergeMap(labels, generateSelectorLabels(util2.SentinelRoleName, rf.Name))
	svc := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Labels:          labels,
			Name:            util2.GetSentinelHeadlessSvc(rf),
			Namespace:       rf.Namespace,
			OwnerReferences: ownerRefs,
		},
		Spec: corev1.ServiceSpec{
			Ports:     []corev1.ServicePort{sentinelPort},
			Selector:  labels,
			ClusterIP: corev1.ClusterIPNone,
		},
	}

	return svc
}

func generateSentinelService(rf *v1alpha1.RedisFailover, labels map[string]string, ownerRefs []metav1.OwnerReference) *corev1.Service {
	name := util2.GetSentinelName(rf)
	namespace := rf.Namespace

	sentinelTargetPort := intstr.FromInt(26379)
	labels = util2.MergeMap(labels, generateSelectorLabels(util2.SentinelRoleName, rf.Name))

	return &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:            name,
			Namespace:       namespace,
			Labels:          labels,
			OwnerReferences: ownerRefs,
		},
		Spec: corev1.ServiceSpec{
			Selector: labels,
			Ports: []corev1.ServicePort{
				{
					Name:       "sentinel",
					Port:       26379,
					TargetPort: sentinelTargetPort,
					Protocol:   "TCP",
				},
			},
		},
	}
}

func generateSentinelConfigMap(rf *v1alpha1.RedisFailover, labels map[string]string, ownerRefs []metav1.OwnerReference) *corev1.ConfigMap {
	name := util2.GetSentinelName(rf)
	namespace := rf.Namespace
	labels = util2.MergeMap(labels, generateSelectorLabels(util2.RedisRoleName, rf.Name))
	sentinelConfigContent := `sentinel monitor mymaster 127.0.0.1 6379 2
sentinel down-after-milliseconds mymaster 1000
sentinel failover-timeout mymaster 3000
sentinel parallel-syncs mymaster 2`
	return &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:            name,
			Namespace:       namespace,
			Labels:          labels,
			OwnerReferences: ownerRefs,
		},
		Data: map[string]string{
			util2.SentinelConfigFileName: sentinelConfigContent,
		},
	}
}

func generateSentinelReadinessProbeConfigMap(rf *v1alpha1.RedisFailover, labels map[string]string, ownerRefs []metav1.OwnerReference) *corev1.ConfigMap {
	name := util2.GetSentinelReadinessConfigmap(rf)
	namespace := rf.Namespace
	labels = util2.MergeMap(labels, generateSelectorLabels(util2.RedisRoleName, rf.Name))
	checkContent := `#!/usr/bin/env sh
set -eou pipefail
redis-cli -h $(hostname) -p 26379 ping
slaves=$(redis-cli -h $(hostname) -p 26379 info sentinel|grep master0| grep -Eo 'slaves=[0-9]+' | awk -F= '{print $2}')
status=$(redis-cli -h $(hostname) -p 26379 info sentinel|grep master0| grep -Eo 'status=\w+' | awk -F= '{print $2}')
if [ "$status" != "ok" ]; then 
    exit 1
fi
if [ $slaves -le 1 ]; then
	exit 1
fi`
	return &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:            name,
			Namespace:       namespace,
			Labels:          labels,
			OwnerReferences: ownerRefs,
		},
		Data: map[string]string{
			"readiness.sh": checkContent,
		},
	}

}

func generateRedisShutdownConfigMap(rf *v1alpha1.RedisFailover, labels map[string]string, ownerRefs []metav1.OwnerReference) *corev1.ConfigMap {
	name := util2.GetRedisShutdownConfigMapName(rf)
	namespace := rf.Namespace
	labels = util2.MergeMap(labels, generateSelectorLabels(util2.RedisRoleName, rf.Name))
	envSentinelHost := fmt.Sprintf("REDIS_SENTINEL_%s_SERVICE_HOST", strings.ToUpper(rf.Name))
	envSentinelPort := fmt.Sprintf("REDIS_SENTINEL_%s_SERVICE_PORT_SENTINEL", strings.ToUpper(rf.Name))
	shutdownContent := fmt.Sprintf(`#!/usr/bin/env sh
master=""
response_code=""
while [ "$master" = "" ]; do
	echo "Asking sentinel who is master..."
	master=$(redis-cli -h ${%s} -p ${%s} --csv SENTINEL get-master-addr-by-name mymaster | tr ',' ' ' | tr -d '\"' |cut -d' ' -f1)
	sleep 1
done
echo "Master is $master, doing redis save..."
redis-cli SAVE
if [ $master = $(hostname -i) ]; then
	while [ ! "$response_code" = "OK" ]; do
  		response_code=$(redis-cli -h ${%s} -p ${%s} SENTINEL failover mymaster)
		echo "after failover with code $response_code"
		sleep 1
	done
fi`, envSentinelHost, envSentinelPort, envSentinelHost, envSentinelPort)

	return &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:            name,
			Namespace:       namespace,
			Labels:          labels,
			OwnerReferences: ownerRefs,
		},
		Data: map[string]string{
			"shutdown.sh": shutdownContent,
		},
	}
}

func generatePodDisruptionBudget(name string, namespace string, labels map[string]string, ownerRefs []metav1.OwnerReference, minAvailable intstr.IntOrString) *policyv1beta1.PodDisruptionBudget {
	return &policyv1beta1.PodDisruptionBudget{
		ObjectMeta: metav1.ObjectMeta{
			Name:            name,
			Namespace:       namespace,
			Labels:          labels,
			OwnerReferences: ownerRefs,
		},
		Spec: policyv1beta1.PodDisruptionBudgetSpec{
			MinAvailable: &minAvailable,
			Selector: &metav1.LabelSelector{
				MatchLabels: labels,
			},
		},
	}
}

func generateSentinelDeployment(rf *v1alpha1.RedisFailover, labels map[string]string, ownerRefs []metav1.OwnerReference) *v1.Deployment {
	name := util2.GetSentinelName(rf)
	namespace := rf.Namespace
	labels = util2.MergeMap(labels, generateSelectorLabels(util2.SentinelRoleName, rf.Name))
	sentinelCommand := getSentinelCommand(rf)

	return &v1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:            name,
			Namespace:       namespace,
			Labels:          labels,
			OwnerReferences: ownerRefs,
		},
		Spec: v1.DeploymentSpec{
			Replicas: &rf.Spec.Sentinel.Replicas,
			Strategy: v1.DeploymentStrategy{
				Type: "RollingUpdate",
			},
			Selector: &metav1.LabelSelector{
				MatchLabels: labels,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels:      labels,
					Annotations: rf.Spec.Sentinel.PodAnnotations,
				},
				Spec: corev1.PodSpec{
					Affinity:         getAffinity(rf.Spec.Sentinel.Affinity, labels),
					Tolerations:      rf.Spec.Sentinel.Tolerations,
					NodeSelector:     rf.Spec.Sentinel.NodeSelector,
					SecurityContext:  getSecurityContext(rf.Spec.Sentinel.SecurityContext),
					ImagePullSecrets: rf.Spec.Sentinel.ImagePullSecrets,
					InitContainers: []corev1.Container{
						{
							Name:            "sentinel-config-copy",
							Image:           rf.Spec.Sentinel.Image,
							ImagePullPolicy: pullPolicy(rf.Spec.Sentinel.ImagePullPolicy),
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:      "sentinel-config",
									MountPath: "/redis",
								},
								{
									Name:      "sentinel-config-writable",
									MountPath: "/redis-writable",
								},
							},
							Command: []string{
								"cp",
								fmt.Sprintf("/redis/%s", util2.SentinelConfigFileName),
								fmt.Sprintf("/redis-writable/%s", util2.SentinelConfigFileName),
							},
							Resources: corev1.ResourceRequirements{
								Limits: corev1.ResourceList{
									corev1.ResourceCPU:    resource.MustParse("10m"),
									corev1.ResourceMemory: resource.MustParse("32Mi"),
								},
								Requests: corev1.ResourceList{
									corev1.ResourceCPU:    resource.MustParse("10m"),
									corev1.ResourceMemory: resource.MustParse("32Mi"),
								},
							},
						},
					},
					Containers: []corev1.Container{
						{
							Name:            "sentinel",
							Image:           rf.Spec.Sentinel.Image,
							ImagePullPolicy: pullPolicy(rf.Spec.Sentinel.ImagePullPolicy),
							Ports: []corev1.ContainerPort{
								{
									Name:          "sentinel",
									ContainerPort: 26379,
									Protocol:      corev1.ProtocolTCP,
								},
							},
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:      "readiness-probe",
									MountPath: "/redis-probe",
								},
								{
									Name:      "sentinel-config-writable",
									MountPath: "/redis",
								},
							},
							Command: sentinelCommand,
							ReadinessProbe: &corev1.Probe{

								PeriodSeconds:    15,
								FailureThreshold: 5,
								TimeoutSeconds:   5,
								Handler: corev1.Handler{
									Exec: &corev1.ExecAction{
										Command: []string{
											"sh",
											"/redis-probe/readiness.sh",
										},
									},
								},
							},
							LivenessProbe: &corev1.Probe{

								TimeoutSeconds: 5,
								Handler: corev1.Handler{
									Exec: &corev1.ExecAction{
										Command: []string{
											"sh",
											"-c",
											"redis-cli -h $(hostname) -p 26379 ping",
										},
									},
								},
							},
							Resources: rf.Spec.Sentinel.Resources,
						},
					},
					Volumes: []corev1.Volume{
						{
							Name: "sentinel-config",
							VolumeSource: corev1.VolumeSource{
								ConfigMap: &corev1.ConfigMapVolumeSource{
									LocalObjectReference: corev1.LocalObjectReference{
										Name: util2.GetSentinelName(rf),
									},
								},
							},
						},
						{
							Name: "readiness-probe",
							VolumeSource: corev1.VolumeSource{
								ConfigMap: &corev1.ConfigMapVolumeSource{
									LocalObjectReference: corev1.LocalObjectReference{
										Name: util2.GetSentinelReadinessConfigmap(rf),
									},
								},
							},
						},
						{
							Name: "sentinel-config-writable",
							VolumeSource: corev1.VolumeSource{
								EmptyDir: &corev1.EmptyDirVolumeSource{},
							},
						},
					},
				},
			},
		},
	}
}

func generateRedisStatefulSet(rf *v1alpha1.RedisFailover, labels map[string]string,
	ownerRefs []metav1.OwnerReference) *v1.StatefulSet {
	name := util2.GetRedisName(rf)
	namespace := rf.Namespace

	spec := rf.Spec
	redisCommand := getRedisCommand(rf)
	labels = util2.MergeMap(labels, generateSelectorLabels(util2.RedisRoleName, rf.Name))
	volumeMounts := getRedisVolumeMounts(rf)
	volumes := getRedisVolumes(rf)

	probeArg := "redis-cli -h $(hostname)"
	if spec.Auth.SecretPath != "" {
		probeArg = fmt.Sprintf("%s -a ${%s} ping", probeArg,redisPasswordEnv)
	} else {
		probeArg = fmt.Sprintf("%s ping", probeArg)
	}

	ss := &v1.StatefulSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:            name,
			Namespace:       namespace,
			Labels:          labels,
			OwnerReferences: ownerRefs,
		},
		Spec: v1.StatefulSetSpec{
			ServiceName: name,
			Replicas:    &spec.Redis.Replicas,
			UpdateStrategy: v1.StatefulSetUpdateStrategy{
				Type: "RollingUpdate",
			},
			Selector: &metav1.LabelSelector{
				MatchLabels: labels,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels:      labels,
					Annotations: rf.Spec.Redis.PodAnnotations,
				},
				Spec: corev1.PodSpec{
					Affinity:         getAffinity(rf.Spec.Redis.Affinity, labels),
					Tolerations:      rf.Spec.Redis.Tolerations,
					NodeSelector:     rf.Spec.Redis.NodeSelector,
					SecurityContext:  getSecurityContext(rf.Spec.Redis.SecurityContext),
					ImagePullSecrets: rf.Spec.Redis.ImagePullSecrets,
					Containers: []corev1.Container{
						{
							Name:            "redis",
							Image:           rf.Spec.Redis.Image,
							ImagePullPolicy: pullPolicy(rf.Spec.Redis.ImagePullPolicy),
							Ports: []corev1.ContainerPort{
								{
									Name:          "redis",
									ContainerPort: 6379,
									Protocol:      corev1.ProtocolTCP,
								},
							},
							VolumeMounts: volumeMounts,
							Command:      redisCommand,
							ReadinessProbe: &corev1.Probe{
								InitialDelaySeconds: graceTime,
								TimeoutSeconds:      5,
								Handler: corev1.Handler{
									Exec: &corev1.ExecAction{
										Command: []string{
											"sh",
											"-c",
											probeArg,
										},
									},
								},
							},
							LivenessProbe: &corev1.Probe{
								InitialDelaySeconds: graceTime,
								TimeoutSeconds:      5,
								Handler: corev1.Handler{
									Exec: &corev1.ExecAction{
										Command: []string{
											"sh",
											"-c",
											probeArg,
										},
									},
								},
							},
							Resources: rf.Spec.Redis.Resources,
							Lifecycle: &corev1.Lifecycle{
								PreStop: &corev1.Handler{
									Exec: &corev1.ExecAction{
										Command: []string{"/bin/sh", "/redis-shutdown/shutdown.sh"},
									},
								},
							},
						},
					},
					Volumes: volumes,
				},
			},
		},
	}

	if rf.Spec.Redis.Storage.PersistentVolumeClaim != nil {
		pvc := rf.Spec.Redis.Storage.PersistentVolumeClaim.DeepCopy()
		if !rf.Spec.Redis.Storage.KeepAfterDeletion {
			// Set an owner reference so the persistent volumes are deleted when the rc is
			pvc.OwnerReferences = ownerRefs
		}
		ss.Spec.VolumeClaimTemplates = []corev1.PersistentVolumeClaim{
			*pvc,
		}

	}
	if rf.Spec.Auth.SecretPath != "" {
		ss.Spec.Template.Spec.Containers[0].Env = append(ss.Spec.Template.Spec.Containers[0].Env, corev1.EnvVar{
			Name: redisPasswordEnv,
			ValueFrom: &corev1.EnvVarSource{
				SecretKeyRef: &corev1.SecretKeySelector{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: rf.Spec.Auth.SecretPath,
					},
					Key: "password",
				},
			},
		})
	}

	if rf.Spec.Redis.Exporter.Enabled {
		exporter := createRedisExporterContainer(rf)
		ss.Spec.Template.Spec.Containers = append(ss.Spec.Template.Spec.Containers, exporter)
	}

	return ss
}

func createRedisExporterContainer(rf *v1alpha1.RedisFailover) corev1.Container {
	container := corev1.Container{
		Name:            exporterContainerName,
		Image:           rf.Spec.Redis.Exporter.Image,
		ImagePullPolicy: pullPolicy(rf.Spec.Redis.Exporter.ImagePullPolicy),
		Env: []corev1.EnvVar{
			{
				Name: "REDIS_ALIAS",
				ValueFrom: &corev1.EnvVarSource{
					FieldRef: &corev1.ObjectFieldSelector{
						FieldPath: "metadata.name",
					},
				},
			},
		},
		Ports: []corev1.ContainerPort{
			{
				Name:          "http-metrics",
				ContainerPort: 9121,
				Protocol:      corev1.ProtocolTCP,
			},
		},
		Resources: corev1.ResourceRequirements{
			Limits: corev1.ResourceList{
				corev1.ResourceCPU:    resource.MustParse("100m"),
				corev1.ResourceMemory: resource.MustParse("200Mi"),
			},
			Requests: corev1.ResourceList{
				corev1.ResourceCPU:    resource.MustParse("50m"),
				corev1.ResourceMemory: resource.MustParse("100Mi"),
			},
		},
	}
	if rf.Spec.Auth.SecretPath != "" {
		container.Env = append(container.Env, corev1.EnvVar{
			Name: redisPasswordEnv,
			ValueFrom: &corev1.EnvVarSource{
				SecretKeyRef: &corev1.SecretKeySelector{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: rf.Spec.Auth.SecretPath,
					},
					Key: "password",
				},
			},
		})
	}
	return container
}

func getRedisCommand(rf *v1alpha1.RedisFailover) []string {
	cmds := []string{
		"redis-server",
		"--slaveof 127.0.0.1 6379",
		"--tcp-keepalive 60",
		"--save 900 1",
		"--save 300 10",
	}
	return cmds
}

func getAffinity(affinity *corev1.Affinity, labels map[string]string) *corev1.Affinity {
	if affinity != nil {
		return affinity
	}

	// Return a SOFT anti-affinity
	return &corev1.Affinity{
		PodAntiAffinity: &corev1.PodAntiAffinity{
			PreferredDuringSchedulingIgnoredDuringExecution: []corev1.WeightedPodAffinityTerm{
				{
					Weight: 100,
					PodAffinityTerm: corev1.PodAffinityTerm{
						TopologyKey: util2.HostnameTopologyKey,
						LabelSelector: &metav1.LabelSelector{
							MatchLabels: labels,
						},
					},
				},
			},
		},
	}
}

func getSecurityContext(secctx *corev1.PodSecurityContext) *corev1.PodSecurityContext {
	if secctx != nil {
		return secctx
	}

	return nil
}

func pullPolicy(specPolicy corev1.PullPolicy) corev1.PullPolicy {
	if specPolicy == "" {
		return corev1.PullAlways
	}
	return specPolicy
}

func getSentinelCommand(rf *v1alpha1.RedisFailover) []string {
	if len(rf.Spec.Sentinel.Command) > 0 {
		return rf.Spec.Sentinel.Command
	}
	return []string{
		"redis-server",
		fmt.Sprintf("/redis/%s", util2.SentinelConfigFileName),
		"--sentinel",
	}
}

func getRedisVolumeMounts(rf *v1alpha1.RedisFailover) []corev1.VolumeMount {
	volumeMounts := []corev1.VolumeMount{
		//{
		//	Name:      redisConfigurationVolumeName,
		//	MountPath: "/redis",
		//},
		{
			Name:      "redis-shutdown-config",
			MountPath: "/redis-shutdown",
		},
		{
			Name:      getRedisDataVolumeName(rf),
			MountPath: "/data",
		},
	}

	return volumeMounts
}

func getRedisDataVolumeName(rf *v1alpha1.RedisFailover) string {
	switch {
	case rf.Spec.Redis.Storage.PersistentVolumeClaim != nil:
		return rf.Spec.Redis.Storage.PersistentVolumeClaim.ObjectMeta.Name
	case rf.Spec.Redis.Storage.EmptyDir != nil:
		return redisStorageVolumeName
	default:
		return redisStorageVolumeName
	}
}

func getRedisVolumes(rf *v1alpha1.RedisFailover) []corev1.Volume {
	shutdownConfigMapName := util2.GetRedisShutdownConfigMapName(rf)

	executeMode := int32(0744)
	volumes := []corev1.Volume{
		{
			Name: redisShutdownConfigurationVolumeName,
			VolumeSource: corev1.VolumeSource{
				ConfigMap: &corev1.ConfigMapVolumeSource{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: shutdownConfigMapName,
					},
					DefaultMode: &executeMode,
				},
			},
		},
	}

	dataVolume := getRedisDataVolume(rf)
	if dataVolume != nil {
		volumes = append(volumes, *dataVolume)
	}

	return volumes
}

func getRedisDataVolume(rf *v1alpha1.RedisFailover) *corev1.Volume {
	// This will find the volumed desired by the user. If no volume defined
	// an EmptyDir will be used by default
	switch {
	case rf.Spec.Redis.Storage.PersistentVolumeClaim != nil:
		return nil
	case rf.Spec.Redis.Storage.EmptyDir != nil:
		return &corev1.Volume{
			Name: redisStorageVolumeName,
			VolumeSource: corev1.VolumeSource{
				EmptyDir: rf.Spec.Redis.Storage.EmptyDir,
			},
		}
	default:
		return &corev1.Volume{
			Name: redisStorageVolumeName,
			VolumeSource: corev1.VolumeSource{
				EmptyDir: &corev1.EmptyDirVolumeSource{},
			},
		}
	}
}
