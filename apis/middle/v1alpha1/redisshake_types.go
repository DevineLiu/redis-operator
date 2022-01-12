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

// RedisShakeSpec defines the desired state of RedisShake
type RedisShakeSpec struct {
	ModelType            *ModelType                    `json:"modelType"`
	SourceInfo           *SourceInfo                   `json:"source"`
	TargetInfo           *TargetInfo                   `json:"target"`
	FilterInfo           *FilterInfo                   `json:"filter,omitempty"`
	FakeTime             int64                         `json:"fakeTime,omitempty"`
	Parallel             int32                         `json:"parallel,omitempty"`
	QBS                  int64                         `json:"qbs,omitempty"`
	ResumeFromBreakPoint bool                          `json:"resumeFromBreakPoint,omitempty"`
	Sender               *Sender                       `json:"sender,omitempty"`
	KeepAlive            int32                         `json:"keepAlive,omitempty"`
	ReplaceHashTag       bool                          `json:"replaceHashTag ,omitempty"`
	BigKeyThreshold      int64                         `json:"bigKeyThreshold,omitempty"`
	Metric               *Metric                       `json:"metric,omitempty"`
	KeyExists            *KeyExistsType                `json:"keyExists,omitempty"`
	Image                string                        `json:"image,omitempty"`
	ImagePullPolicy      corev1.PullPolicy             `json:"imagePullPolicy,omitempty"`
	Replicas             int32                         `json:"replicas,omitempty"`
	Resources            corev1.ResourceRequirements   `json:"resources,omitempty"`
	Affinity             *corev1.Affinity              `json:"affinity,omitempty"`
	NodeSelector         map[string]string             `json:"nodeSelector,omitempty"`
	SecurityContext      *corev1.PodSecurityContext    `json:"securityContext,omitempty"`
	ImagePullSecrets     []corev1.LocalObjectReference `json:"imagePullSecrets,omitempty"`
	Tolerations          []corev1.Toleration           `json:"tolerations,omitempty"`
	PodAnnotations       map[string]string             `json:"podAnnotations,omitempty"`
	ServiceAnnotations   map[string]string             `json:"serviceAnnotations,omitempty"`
}

type ModelType string

const (
	Decode  ModelType = "decode"
	Restore ModelType = "restore"
	Sync    ModelType = "sync"
	rump    ModelType = "rump"
)

type ScanInfo struct {
	KeyNumber    int64  `json:"keyNumber,omitempty"`
	SpecialCloud string `json:"specialCloud,omitempty"`
	KeyFile      string `json:"keyFile,omitempty"`
}

type Sender struct {
	Size             int64 `json:"size,omitempty"`
	Count            int64 `json:"count,omitempty"`
	DelayChannelSize int64 `json:"delayChannelSize,omitempty"`
}

type Metric struct {
	Enable   bool `json:"enable,omitempty"`
	PrintLog bool `json:"printLog,omitempty"`
}

type KeyExistsType string

const (
	Rewrite KeyExistsType = "rewrite"
	NOne    KeyExistsType = "none"
	Ignore  KeyExistsType = "ignore"
)

type SourceType string

const (
	Standalone SourceType = "standalone"
	Sentinel   SourceType = "sentinel"
	Cluster    SourceType = "cluster"
	Proxy      SourceType = "proxy"
)

type LogLevel string

const (
	None  LogLevel = "none"
	Error LogLevel = "error"
	Warn  LogLevel = "warn"
	Info  LogLevel = "info"
	Debug LogLevel = "debug"
)

type FilterDB struct {
	WhiteList []string `json:"whiteList,omitempty"`
	BlackList []string `json:"blackList,omitempty"`
}

type FilterKey struct {
	WhiteList []string `json:"whiteList,omitempty"`
	BlackList []string `json:"blackList,omitempty"`
}

type FilterCommand struct {
	WhiteList []string `json:"whiteList,omitempty"`
	BlackList []string `json:"blackList,omitempty"`
}

type FilterInfo struct {
	DB      *FilterDB      `json:"db,omitempty"`
	Key     *FilterKey     `json:"key,omitempty"`
	Slot    []string       `json:"slot,omitempty"`
	Command *FilterCommand `json:"command,omitempty"`
	Lua     bool           `json:"lua,omitempty"`
}

type SourceInfo struct {
	Type           SourceType `json:"type"`
	Address        []string   `json:"address,omitempty"`
	PasswordRaw    string     `json:"passwordRaw,omitempty"`
	TlsEnable      bool       `json:"tlsEnable,omitempty"`
	TlsSkipVerify  bool       `json:"tlsSkipVerify,omitempty"`
	RDB            *RDBInput  `json:"rdb,omitempty"`
	ClusterName    string     `json:"clusterName,omitempty"`
	PasswordSecret string     `json:"passwordSecret,omitempty"`
}

type RDBInput struct {
	Input        []string `json:"input,omitempty"`
	Parallel     int32    `json:"parallel,omitempty"`
	SpecialCloud string   `json:"specialCloud,omitempty"`
}

type RDBOutput struct {
	Output  string `json:"output,omitempty"`
	Version string `json:"version,omitempty"`
}

type TargetInfo struct {
	Type           SourceType       `json:"type"`
	Address        []string         `json:"address,omitempty"`
	PasswordRaw    string           `json:"passwordRaw,omitempty"`
	DB             int32            `json:"db,omitempty"`
	DbMap          map[string]int32 `json:"dbmap,omitempty"`
	TlsEnable      bool             `json:"tlsEnable,omitempty"`
	TlsSkipVerify  bool             `json:"tlsSkipVerify,omitempty"`
	ClusterName    string           `json:"clusterName,omitempty"`
	PasswordSecret string           `json:"passwordSecret,omitempty"`
}

type LogInfo struct {
	File  string   `json:"file"`
	Level LogLevel `json:"level"`
}

// RedisShakeStatus defines the observed state of RedisShake
type RedisShakeStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// RedisShake is the Schema for the redisshakes API
type RedisShake struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   RedisShakeSpec   `json:"spec,omitempty"`
	Status RedisShakeStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// RedisShakeList contains a list of RedisShake
type RedisShakeList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []RedisShake `json:"items"`
}

func init() {
	SchemeBuilder.Register(&RedisShake{}, &RedisShakeList{})
}
