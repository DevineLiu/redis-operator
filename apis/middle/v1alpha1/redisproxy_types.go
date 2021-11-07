/*
Copyright 2021.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package v1alpha1

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// RedisProxySpec defines the desired state of RedisProxy
type RedisProxySpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Foo is an example field of RedisProxy. Edit redisproxy_types.go to remove/update
	ProxyInfo          ProxyInfo                     `json:"proxyInfo"`
	Image              string                        `json:"image,omitempty"`
	ImagePullPolicy    corev1.PullPolicy             `json:"imagePullPolicy,omitempty"`
	Replicas           int32                         `json:"replicas,omitempty"`
	Resources          corev1.ResourceRequirements   `json:"resources,omitempty"`
	Affinity           *corev1.Affinity              `json:"affinity,omitempty"`
	SecurityContext    *corev1.PodSecurityContext    `json:"securityContext,omitempty"`
	ImagePullSecrets   []corev1.LocalObjectReference `json:"imagePullSecrets,omitempty"`
	Auth               AuthSettings                  `json:"auth,omitempty"`
	Tolerations        []corev1.Toleration           `json:"tolerations,omitempty"`
	NodeSelector       map[string]string             `json:"nodeSelector,omitempty"`
	PodAnnotations     map[string]string             `json:"podAnnotations,omitempty"`
	ServiceAnnotations map[string]string             `json:"serviceAnnotations,omitempty"`
}

type ProxyInfo struct {
	Architecture  string `json:"architecture"`
	InstanceName  string `json:"instanceName"`
	WorkerThreads int32  `json:"workThreads,omitempty"`
	ClientTimeout int32  `json:"clientTimeout,omitempty"`
}

// RedisProxyStatus defines the observed state of RedisProxy
type RedisProxyStatus struct {
	Conditions []Condition `json:"conditions,omitempty"`
	Version    string      `json:"version,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// RedisProxy is the Schema for the redisproxies API
type RedisProxy struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   RedisProxySpec   `json:"spec,omitempty"`
	Status RedisProxyStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// RedisProxyList contains a list of RedisProxy
type RedisProxyList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []RedisProxy `json:"items"`
}

func init() {
	SchemeBuilder.Register(&RedisProxy{}, &RedisProxyList{})
}
