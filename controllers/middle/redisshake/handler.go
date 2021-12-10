package redisshake

import (
	"context"
	"fmt"

	middlev1alpha1 "github.com/DevineLiu/redis-operator/apis/middle/v1alpha1"
	"github.com/DevineLiu/redis-operator/controllers/middle/client/k8s"
	"github.com/DevineLiu/redis-operator/controllers/middle/shakeservice"
	"github.com/DevineLiu/redis-operator/controllers/util"
	"github.com/go-logr/logr"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type RedisShakeHandler struct {
	Logger       logr.Logger
	Record       record.EventRecorder
	K8sService   k8s.Services
	RsServices   shakeservice.RedisShakeClient
	StatusWriter StatusWriter
}

func (r *RedisShakeHandler) Do(rs *middlev1alpha1.RedisShake) error {
	if err := rs.Validate(); err != nil {
		r.Record.Event(rs, v1.EventTypeWarning, "Valiadte", fmt.Sprintf("err: %s", err.Error()))
		return err
	}
	oRefs := r.createOwnerReferences(rs)
	labels := r.getLabels(rs)
	r.Logger.WithValues("namespace", rs.Namespace, "name", rs.Name).V(2).Info("Ensure redisShake...")
	r.Record.Event(rs, v1.EventTypeNormal, "Ensure", "Ensure redisShake running")
	if err := r.Ensure(rs, labels, oRefs); err != nil {
		return err
	}
	return nil
}

func (r *RedisShakeHandler) createOwnerReferences(rs *middlev1alpha1.RedisShake) []metav1.OwnerReference {
	rcvk := middlev1alpha1.GroupVersion.WithKind(rs.Kind)
	return []metav1.OwnerReference{
		*metav1.NewControllerRef(rs, rcvk),
	}
}

func (r *RedisShakeHandler) getLabels(rs *middlev1alpha1.RedisShake) map[string]string {
	dynLabels := map[string]string{
		"redis/v1beta1": fmt.Sprintf("%s%c%s", rs.Namespace, '_', rs.Name),
	}
	defaultLabels := map[string]string{
		"redis/managed-by": "redis-shake",
	}

	return util.MergeMap(defaultLabels, dynLabels, rs.Labels)
}

type StatusWriter struct {
	client.Client
	Ctx context.Context
}

type StatusWrite interface {
	Update(rf *middlev1alpha1.RedisShake, opts ...client.UpdateOption) error
	Patch(rf *middlev1alpha1.RedisShake, patch client.Patch, opts ...client.PatchOption) error
}

func (s *StatusWriter) Patch(rf *middlev1alpha1.RedisShake, patch client.Patch, opts ...client.PatchOption) error {
	err := s.Status().Patch(s.Ctx, rf, patch, opts...)
	return err
}

func (s *StatusWriter) Update(rf *middlev1alpha1.RedisShake, opts ...client.UpdateOption) error {
	err := s.Status().Update(s.Ctx, rf, opts...)
	return err
}
