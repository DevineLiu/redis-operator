package proxyservice

import (
	"fmt"
	middlev1alpha1 "github.com/DevineLiu/redis-operator/apis/middle/v1alpha1"
	util2 "github.com/DevineLiu/redis-operator/controllers/util"
	v1 "k8s.io/api/apps/v1"

	corev1 "k8s.io/api/core/v1"
	policyv1beta1 "k8s.io/api/policy/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)


func generateRedisProxyDeployment(rp *middlev1alpha1.RedisProxy, labels map[string]string, ownrf []metav1.OwnerReference)  *v1.Deployment {
	name := util2.GetRedisProxyName(rp)
	namespace := rp.Namespace
	labels = util2.MergeMap(labels, generateSelectorLabels(util2.ProxyRoleName, rp.Name))
	return &v1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:            name,
			Namespace:       namespace,
			Labels:          labels,
			OwnerReferences: ownrf,
		},
		Spec: v1.DeploymentSpec{
			Replicas: &rp.Spec.Replicas,
			Strategy: v1.DeploymentStrategy{
				Type: "RollingUpdate",
			},
			Selector: &metav1.LabelSelector{
				MatchLabels: labels,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels:      labels,
					Annotations: rp.Spec.PodAnnotations,
				},
				Spec: corev1.PodSpec{
					Affinity:         getAffinity(rp.Spec.Affinity, labels),
					Tolerations:      rp.Spec.Tolerations,
					NodeSelector:     rp.Spec.NodeSelector,
					SecurityContext:  getSecurityContext(rp.Spec.SecurityContext),
					ImagePullSecrets: rp.Spec.ImagePullSecrets,
					Containers: []corev1.Container{
						{
							Name:            "proxy",
							Image:           rp.Spec.Image,
							ImagePullPolicy: pullPolicy(rp.Spec.ImagePullPolicy),
							Ports: []corev1.ContainerPort{
								{
									Name:          "redis-proxy",
									ContainerPort: 6379,
									Protocol:      corev1.ProtocolTCP,
								},
							},
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:      "proxy-config",
									MountPath: "/predixy",
								},
							},
							Command: getProxyCommand(),
							Resources: rp.Spec.Resources,
						},
					},
					Volumes: []corev1.Volume{
						{
							Name: "proxy-config",
							VolumeSource: corev1.VolumeSource{
								ConfigMap: &corev1.ConfigMapVolumeSource{
									LocalObjectReference: corev1.LocalObjectReference{
										Name: util2.GetRedisProxyName(rp),
									},
								},
							},
						},
					},
				},
			},
		},
	}

}

func generateRedisProxyService(rp *middlev1alpha1.RedisProxy, labels map[string]string, ownrf []metav1.OwnerReference) *corev1.Service {
	name := util2.GetRedisProxyName(rp)
	namespace := rp.Namespace
	proxyTargetPort := intstr.FromInt(6379)
	labels = util2.MergeMap(labels, generateSelectorLabels(util2.ProxyRoleName, rp.Name))

	return &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:            name,
			Namespace:       namespace,
			Labels:          labels,
			OwnerReferences: ownrf,
		},
		Spec: corev1.ServiceSpec{
			Selector: labels,
			Ports: []corev1.ServicePort{
				{
					Name:       "redis",
					Port:       6379,
					TargetPort: proxyTargetPort,
					Protocol:   "TCP",
				},
			},
		},
	}
}

func generateRedisProxyNodePortService(rp *middlev1alpha1.RedisProxy, labels map[string]string, ownrf []metav1.OwnerReference) *corev1.Service {
	name := util2.GetRedisProxyNodePortName(rp)
	namespace := rp.Namespace
	sentinelTargetPort := intstr.FromInt(6379)
	labels = util2.MergeMap(labels, generateSelectorLabels(util2.ProxyRoleName, rp.Name))

	return &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:            name,
			Namespace:       namespace,
			Labels:          labels,
			OwnerReferences: ownrf,
		},
		Spec: corev1.ServiceSpec{
			Selector: labels,
			Type:      corev1.ServiceTypeNodePort,
			Ports: []corev1.ServicePort{
				{
					Name:       "redis",
					Port:       6379,
					TargetPort: sentinelTargetPort,
					Protocol:   "TCP",

				},
			},
		},
	}
}

func generateRedisProxyConfigMap(rp *middlev1alpha1.RedisProxy, labels map[string]string, ownrf []metav1.OwnerReference,password string) *corev1.ConfigMap {
	name := util2.GetRedisProxyName(rp)
	namespace := rp.Namespace
	labels = util2.MergeMap(labels, generateSelectorLabels(util2.ProxyName, rp.Name))
	ProxyConfigMapContent := `
LogVerbSample 0
LogDebugSample 0
LogInfoSample 0
LogNoticeSample 0
LogWarnSample 1
LogErrorSample 1
Bind 0.0.0.0:6379

WorkerThreads %d
Name %s
ClientTimeout %d

Authority {
    Auth %s {
        Mode admin
    }
}

ClusterServerPool {
    Password %s
    Servers {
        + %s-0:6379
        + %s-1:6379
        + %s-2:6379
    }
}
`

	ConfigMapContent := fmt.Sprintf(ProxyConfigMapContent,rp.Spec.ProxyInfo.WorkerThreads,rp.Name,
	rp.Spec.ProxyInfo.ClientTimeout,password,password, rp.Spec.ProxyInfo.InstanceName,
	rp.Spec.ProxyInfo.InstanceName,rp.Spec.ProxyInfo.InstanceName)
	return &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:            name,
			Namespace:       namespace,
			Labels:          labels,
			OwnerReferences: ownrf,
		},
		Data: map[string]string{
			util2.ProxyConfigFileName: ConfigMapContent,
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

func getProxyCommand() []string {
	return []string{
		"predixy",
		fmt.Sprintf("/predixy/%s", util2.ProxyConfigFileName),
	}
}