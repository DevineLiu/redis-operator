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
	panic("impl me")
}

func generateRedisProxyService(rp *middlev1alpha1.RedisProxy, labels map[string]string, ownrf []metav1.OwnerReference) *corev1.Service {
	name := util2.GetRedisProxyName(rp)
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
			ClusterIP: corev1.ClusterIPNone,
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

WorkerThreads %d
Name %s
ClientTimeout %d

Authority {
    Auth %s {
        Mode admin
    }
}

ClusterServerPool {
    Servers {
        + %s-0:6379
        + %s-1:6379
        + %s-2:6379
    }
}
`

	ConfigMapContent := fmt.Sprintf(ProxyConfigMapContent,rp.Spec.ProxyInfo.WorkerThreads,rp.Name,
	rp.Spec.ProxyInfo.ClientTimeout,password, rp.Spec.ProxyInfo.InstanceName,
	rp.Spec.ProxyInfo.InstanceName,rp.Spec.ProxyInfo.InstanceName)
	return &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:            name,
			Namespace:       namespace,
			Labels:          labels,
			OwnerReferences: ownrf,
		},
		Data: map[string]string{
			"proxy.conf": ConfigMapContent,
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