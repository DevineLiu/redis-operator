/*


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

package v1

import (
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// RedisBackupSpec defines the desired state of RedisBackup
type RedisBackupSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	Source  RedisBackupSource `json:"source,omitempty"`
	Storage resource.Quantity `json:"storage,omitempty"`
	Image   string            `json:"image,omitempty"`
}

// RedisBackupStatus defines the observed state of RedisBackup
type RedisBackupStatus struct {
	JobName        string       `json:"jobName,omitempty"`
	CompletionTime *metav1.Time `json:"completionTime,omitempty"`
	// +optional
	// where store backup data in
	Destination   string               `json:"destination,omitempty"`
	Condition     RedisBackupCondition `json:"condition,omitempty"`
	LastCheckTime *metav1.Time         `json:"lastCheckTime,omitempty"`
	// +optional
	// show message when backup fail
	Message string `json:"message,omitempty"`
}

type RedisBackupCondition string

// These are valid conditions of a redis backup.
const (
	// RedisBackupRunning means the job running its execution.
	RedisBackupRunning RedisBackupCondition = "Running"
	// RedisBackupComplete means the job has completed its execution.
	RedisBackupComplete RedisBackupCondition = "Complete"
	// RedisBackupFailed means the job has failed its execution.
	RedisBackupFailed RedisBackupCondition = "Failed"
)

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// RedisBackup is the Schema for the redisbackups API
type RedisBackup struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   RedisBackupSpec   `json:"spec,omitempty"`
	Status RedisBackupStatus `json:"status,omitempty"`
}

type RedisBackupSource struct {
	RedisFailoverName string `json:"redisFailoverName"`
	StorageClassName  string `json:"storageClassName"`
}

// +kubebuilder:object:root=true

// RedisBackupList contains a list of RedisBackup
type RedisBackupList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []RedisBackup `json:"items"`
}

func init() {
	SchemeBuilder.Register(&RedisBackup{}, &RedisBackupList{})
}
