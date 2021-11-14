package v1alpha1

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sort"
	"time"
)

// Phase of the RF status
type Phase string

// Condition saves the state information of the redis RedisFailover
type Condition struct {
	// Status of RedisFailover condition.
	Type ConditionType `json:"type"`
	// Status of the condition, one of True, False, Unknown.
	Status corev1.ConditionStatus `json:"status"`
	// The last time this condition was updated.
	LastUpdateTime string      `json:"lastUpdateTime,omitempty"`
	LastUpdateAt   metav1.Time `json:"-"`
	// Last time the condition transitioned from one status to another.
	LastTransitionTime string `json:"lastTransitionTime,omitempty"`
	// The reason for the condition's last transition.
	Reason string `json:"reason,omitempty"`
	// A human readable message indicating details about the transition.
	Message string `json:"message,omitempty"`
}

// ConditionType defines the condition that the RF can have
type ConditionType string

const (
	RedisFailoverConditionAvailable   ConditionType = "Available"
	RedisFailoverConditionHealthy     ConditionType = "Healthy"
	RedisFailoverConditionRunning                   = "Running"
	RedisFailoverConditionCreating                  = "Creating"
	RedisFailoverConditionRecovering                = "Recovering"
	RedisFailoverConditionScaling                   = "Scaling"
	RedisFailoverConditionScalingDown               = "ScalingDown"
	RedisFailoverConditionUpgrading                 = "Upgrading"
	RedisFailoverConditionUpdating                  = "Updating"
	RedisFailoverConditionFailed                    = "Failed"
)

func (rf *RedisFailoverStatus) DescConditionsByTime() {
	sort.Slice(rf.Conditions, func(i, j int) bool {
		return rf.Conditions[i].LastUpdateAt.After(rf.Conditions[j].LastUpdateAt.Time)
	})
}

func (rp *RedisProxyStatus) DescConditionsByTime() {
	sort.Slice(rp.Conditions, func(i, j int) bool {
		return rp.Conditions[i].LastUpdateAt.After(rp.Conditions[j].LastUpdateAt.Time)
	})
}

func (rf *RedisFailoverStatus) SetScalingUpCondition(message string) {
	c := newRedisFailoverCondition(RedisFailoverConditionScaling, corev1.ConditionTrue, "Scaling up", message)
	rf.setRedisFailoverCondition(*c)
}

func (rf *RedisFailoverStatus) SetCreateCondition(message string) {
	c := newRedisFailoverCondition(RedisFailoverConditionCreating, corev1.ConditionTrue, "Creating", message)
	rf.setRedisFailoverCondition(*c)
}

func (rf *RedisFailoverStatus) SetScalingDownCondition(message string) {
	c := newRedisFailoverCondition(RedisFailoverConditionScaling, corev1.ConditionTrue, "Scaling down", message)
	rf.setRedisFailoverCondition(*c)
}

func (rf *RedisFailoverStatus) SetWaitingPodReadyCondition(message string) {
	c := newRedisFailoverCondition(RedisFailoverConditionScaling, corev1.ConditionTrue, "WaitingPod", message)
	rf.setRedisFailoverCondition(*c)

}

func (rf *RedisFailoverStatus) IsLastConditionWaitingPodReady() bool {
	if len(rf.Conditions) > 0 {
		rf.DescConditionsByTime()

		codition := rf.Conditions[0]
		return codition.Reason == "WaitingPod"
	}
	return false
}

func (rp *RedisProxyStatus) IsLastConditionUpgrading() bool {
	if len(rp.Conditions) > 0 {
		rp.DescConditionsByTime()

		codition := rp.Conditions[0]
		return codition.Reason == "RedisProxy Upgrading"
	}
	return false
}



func (rf *RedisFailoverStatus) SetUpgradingCondition(message string) {
	c := newRedisFailoverCondition(RedisFailoverConditionUpgrading, corev1.ConditionTrue,
		"RedisFailover upgrading", message)
	rf.setRedisFailoverCondition(*c)
}

func (rp *RedisProxyStatus) SetUpgradingCondition(message string) {
	c := newRedisFailoverCondition(RedisFailoverConditionUpgrading, corev1.ConditionTrue,
		"RedisProxy Upgrading", message)
	rp.setRedisProxyCondition(*c)
}


func (rf *RedisFailoverStatus) SetUpdatingCondition(message string) {
	c := newRedisFailoverCondition(RedisFailoverConditionUpdating, corev1.ConditionTrue,
		"RedisFailover updating", message)
	rf.setRedisFailoverCondition(*c)
}

func (rf *RedisFailoverStatus) SetReadyCondition(message string) {
	c := newRedisFailoverCondition(RedisFailoverConditionHealthy, corev1.ConditionTrue, "RedisFailover available", message)
	rf.setRedisFailoverCondition(*c)
}

func (rp *RedisProxyStatus) SetReadyCondition(message string) {
	c := newRedisFailoverCondition(RedisFailoverConditionHealthy, corev1.ConditionTrue, "Redis Ready", message)
	rp.setRedisProxyCondition(*c)
}

func (rf *RedisFailoverStatus) SetFailedCondition(message string) {
	c := newRedisFailoverCondition(RedisFailoverConditionFailed, corev1.ConditionTrue,
		"RedisFailover failed", message)
	rf.setRedisFailoverCondition(*c)
}

func (rf *RedisFailoverStatus) ClearCondition(t ConditionType) {
	pos, _ := getRedisFailoverCondition(rf, t)
	if pos == -1 {
		return
	}
	rf.Conditions = append(rf.Conditions[:pos], rf.Conditions[pos+1:]...)
}

func (rf *RedisFailoverStatus) setRedisFailoverCondition(c Condition) {
	pos, cp := getRedisFailoverCondition(rf, c.Type)
	if cp != nil &&
		cp.Status == c.Status && cp.Reason == c.Reason && cp.Message == c.Message {
		now := time.Now()
		nowString := now.Format(time.RFC3339)
		rf.Conditions[pos].LastUpdateAt = metav1.Time{Time: now}
		rf.Conditions[pos].LastUpdateTime = nowString
		return
	}

	if cp != nil {
		rf.Conditions[pos] = c
	} else {
		rf.Conditions = append(rf.Conditions, c)
	}
}

func (rp *RedisProxyStatus) setRedisProxyCondition(c Condition) {
	pos, cp := getRedisProxyCondition(rp, c.Type)
	if cp != nil &&
		cp.Status == c.Status && cp.Reason == c.Reason && cp.Message == c.Message {
		now := time.Now()
		nowString := now.Format(time.RFC3339)
		rp.Conditions[pos].LastUpdateAt = metav1.Time{Time: now}
		rp.Conditions[pos].LastUpdateTime = nowString
		return
	}

	if cp != nil {
		rp.Conditions[pos] = c
	} else {
		rp.Conditions = append(rp.Conditions, c)
	}
}


func getRedisFailoverCondition(status *RedisFailoverStatus, t ConditionType) (int, *Condition) {
	for i, c := range status.Conditions {
		if t == c.Type {
			return i, &c
		}
	}
	return -1, nil
}

func getRedisProxyCondition(status *RedisProxyStatus, t ConditionType) (int, *Condition) {
	for i, c := range status.Conditions {
		if t == c.Type {
			return i, &c
		}
	}
	return -1, nil
}


func newRedisFailoverCondition(condType ConditionType, status corev1.ConditionStatus, reason, message string) *Condition {
	now := time.Now()
	nowString := now.Format(time.RFC3339)
	return &Condition{
		Type:               condType,
		Status:             status,
		LastUpdateTime:     nowString,
		LastUpdateAt:       metav1.Time{Time: now},
		LastTransitionTime: nowString,
		Reason:             reason,
		Message:            message,
	}
}
