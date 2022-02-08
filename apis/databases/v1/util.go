package v1

const (
	Fail            Phase = "Fail"
	Creating        Phase = "Creating"
	Pending         Phase = "Pending"
	Ready           Phase = "Ready"
	WaitingPodReady Phase = "WaitingPodReady"
)

func (r *RedisFailoverStatus) SetFailedPhase(message string) {
	r.Phase = Fail
	r.Message = message
}

func (r *RedisFailoverStatus) SetWaitingPodReady(message string) {
	r.Phase = WaitingPodReady
	r.Message = message
}

func (r *RedisFailoverStatus) IsWaitingPodReady() bool {
	return r.Phase == WaitingPodReady
}

func (r *RedisFailoverStatus) SetReady(message string) {
	r.Phase = Ready
	r.Message = message
}
