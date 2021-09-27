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
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// RedisFailoverSpec defines the desired state of RedisFailover
type RedisFailoverSpec struct {
	Redis          RedisSettings    `json:"redis,omitempty"`
	Sentinel       SentinelSettings `json:"sentinel,omitempty"`
	Auth           AuthSettings     `json:"auth,omitempty"`
	LabelWhitelist []string         `json:"labelWhitelist,omitempty"`
}

// RedisCommandRename defines the specification of a "rename-command" configuration option
type RedisCommandRename struct {
	From string `json:"from,omitempty"`
	To   string `json:"to,omitempty"`
}

// RedisSettings defines the specification of the redis cluster
type RedisSettings struct {
	Image                string                        `json:"image,omitempty"`
	ImagePullPolicy      corev1.PullPolicy             `json:"imagePullPolicy,omitempty"`
	Replicas             int32                         `json:"replicas,omitempty"`
	Resources            corev1.ResourceRequirements   `json:"resources,omitempty"`
	ConfigConfigMap      string                        `json:"configConfigMap,omitempty"`
	CustomConfig         map[string]string             `json:"customConfig,omitempty"`
	CustomCommandRenames []RedisCommandRename          `json:"customCommandRenames,omitempty"`
	Command              []string                      `json:"command,omitempty"`
	ShutdownConfigMap    string                        `json:"shutdownConfigMap,omitempty"`
	Storage              RedisStorage                  `json:"storage,omitempty"`
	Exporter             RedisExporter                 `json:"exporter,omitempty"`
	Affinity             *corev1.Affinity              `json:"affinity,omitempty"`
	SecurityContext      *corev1.PodSecurityContext    `json:"securityContext,omitempty"`
	ImagePullSecrets     []corev1.LocalObjectReference `json:"imagePullSecrets,omitempty"`
	Tolerations          []corev1.Toleration           `json:"tolerations,omitempty"`
	NodeSelector         map[string]string             `json:"nodeSelector,omitempty"`
	PodAnnotations       map[string]string             `json:"podAnnotations,omitempty"`
	ServiceAnnotations   map[string]string             `json:"serviceAnnotations,omitempty"`
	HostNetwork          bool                          `json:"hostNetwork,omitempty"`
	DNSPolicy            corev1.DNSPolicy              `json:"dnsPolicy,omitempty"`
	Backup               RedisBackup                   `json:"backup,omitempty"`
	Restore              RedisRestore                  `json:"restore,omitempty"`
}

// SentinelSettings defines the specification of the sentinel cluster
type SentinelSettings struct {
	Image              string                        `json:"image,omitempty"`
	ImagePullPolicy    corev1.PullPolicy             `json:"imagePullPolicy,omitempty"`
	Replicas           int32                         `json:"replicas,omitempty"`
	Resources          corev1.ResourceRequirements   `json:"resources,omitempty"`
	CustomConfig       []string                      `json:"customConfig,omitempty"`
	Command            []string                      `json:"command,omitempty"`
	Affinity           *corev1.Affinity              `json:"affinity,omitempty"`
	SecurityContext    *corev1.PodSecurityContext    `json:"securityContext,omitempty"`
	ImagePullSecrets   []corev1.LocalObjectReference `json:"imagePullSecrets,omitempty"`
	Tolerations        []corev1.Toleration           `json:"tolerations,omitempty"`
	NodeSelector       map[string]string             `json:"nodeSelector,omitempty"`
	PodAnnotations     map[string]string             `json:"podAnnotations,omitempty"`
	ServiceAnnotations map[string]string             `json:"serviceAnnotations,omitempty"`
	Exporter           SentinelExporter              `json:"exporter,omitempty"`
	HostNetwork        bool                          `json:"hostNetwork,omitempty"`
	DNSPolicy          corev1.DNSPolicy              `json:"dnsPolicy,omitempty"`
}

// AuthSettings contains settings about auth
type AuthSettings struct {
	SecretPath string `json:"secretPath,omitempty"`
}

// RedisExporter defines the specification for the redis exporter
type RedisExporter struct {
	Enabled         bool              `json:"enabled,omitempty"`
	Image           string            `json:"image,omitempty"`
	ImagePullPolicy corev1.PullPolicy `json:"imagePullPolicy,omitempty"`
}

// SentinelExporter defines the specification for the sentinel exporter
type SentinelExporter struct {
	Enabled         bool              `json:"enabled,omitempty"`
	Image           string            `json:"image,omitempty"`
	ImagePullPolicy corev1.PullPolicy `json:"imagePullPolicy,omitempty"`
}

// RedisStorage defines the structure used to store the Redis Data

type RedisStorage struct {
	KeepAfterDeletion bool                         `json:"keepAfterDeletion,omitempty"`
	EmptyDir          *corev1.EmptyDirVolumeSource `json:"emptyDir,omitempty"`

	PersistentVolumeClaim *corev1.PersistentVolumeClaim `json:"persistentVolumeClaim,omitempty"`
}

// RedisBackup defines the structure used to backup the Redis Data
type RedisBackupSetting struct {
	Image    string     `json:"image,omitempty"`
	Schedule []Schedule `json:"schedule,omitempty"`
}

type Schedule struct {
	Name              string             `json:"name"`
	Schedule          string             `json:"schedule"`
	Keep              int32              `json:"keep"`
	KeepAfterDeletion bool               `json:"keepAfterDeletion,omitempty"`
	Storage           RedisBackupStorage `json:"storage"`
}

type RedisBackupStorage struct {
	StorageClassName string            `json:"storageClassName,omitempty"`
	Size             resource.Quantity `json:"size,omitempty"`
}

// RedisBackup defines the structure used to restore the Redis Data
type RedisRestore struct {
	Image           string            `json:"image,omitempty"`
	ImagePullPolicy corev1.PullPolicy `json:"imagePullPolicy,omitempty"`
	BackupName      string            `json:"backupName,omitempty"`
}

// RedisStatus
type RedisFailoverStatus struct {
	Conditions []Condition `json:"conditions,omitempty"`
	// The last time this condition was updated.
	LastUpdateTime metav1.Time `json:"lastUpdateTime,omitempty"`
	// Last time the condition transitioned from one status to another.
	LastTransitionTime metav1.Time `json:"lastTransitionTime,omitempty"`
	Message            string      `json:"message,omitempty"`
	// Creating, Pending, Fail, Ready
	Phase string `json:"phase,omitempty"`

	Instance RedisStatusInstance `json:"instance,omitempty"`
	Master   RedisStatusMaster   `json:"master,omitempty"`
	Version  string              `json:"version,omitempty"`
}

type RedisStatusInstance struct {
	Redis    RedisStatusInstanceRedis    `json:"redis,omitempty"`
	Sentinel RedisStatusInstanceSentinel `json:"sentinel,omitempty"`
}

type RedisStatusInstanceRedis struct {
	Size  int32 `json:"size,omitempty"`
	Ready int32 `json:"ready,omitempty"`
}

type RedisStatusInstanceSentinel struct {
	Size      int32  `json:"size,omitempty"`
	Ready     int32  `json:"ready,omitempty"`
	Service   string `json:"service,omitempty"`
	ClusterIP string `json:"clusterIp,omitempty"`
	Port      string `json:"port,omitempty"`
}

type RedisStatusMaster struct {
	Name    string                  `json:"name"`
	Status  RedisStatusMasterStatus `json:"status"`
	Address string                  `json:"address"`
}

type RedisStatusMasterStatus string

const (
	RedisStatusMasterOK   RedisStatusMasterStatus = "ok"
	RedisStatusMasterDown RedisStatusMasterStatus = "down"
)

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// RedisFailover is the Schema for the redisfailovers API
type RedisFailover struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   RedisFailoverSpec   `json:"spec,omitempty"`
	Status RedisFailoverStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// RedisFailoverList contains a list of RedisFailover
type RedisFailoverList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []RedisFailover `json:"items"`
}

func init() {
	SchemeBuilder.Register(&RedisFailover{}, &RedisFailoverList{})
}
