package redisproxy

import (
	"fmt"
	"github.com/DevineLiu/redis-operator/controllers/middle/proxyservice"

	middlev1alpha1 "github.com/DevineLiu/redis-operator/apis/middle/v1alpha1"
	"github.com/DevineLiu/redis-operator/controllers/middle/client/k8s"
	"github.com/DevineLiu/redis-operator/controllers/util"
	"github.com/go-logr/logr"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/record"
)

type RedisProxyHandler struct {
	Logger     logr.Logger
	Record     record.EventRecorder
	K8sService k8s.Services
	RpServices proxyservice.RedisProxyClient
}

func (r *RedisProxyHandler) Do(rp *middlev1alpha1.RedisProxy) error {

	if err := rp.Validate(); err != nil {
		r.Record.Event(rp, v1.EventTypeWarning, "Valiadte", fmt.Sprintf("err: %s", err.Error()))
		return err
	}

	oRefs := r.createOwnerReferences(rp)
	labels := r.getLabels(rp)
	r.Logger.WithValues("namespace", rp.Namespace, "name", rp.Name).V(2).Info("Ensure...")
	r.Record.Event(rp, v1.EventTypeNormal, "Ensure", "Ensure running")
	if err := r.Ensure(rp, labels, oRefs); err != nil {
		return nil
	}

	return nil
}

func (r *RedisProxyHandler) createOwnerReferences(rp *middlev1alpha1.RedisProxy) []metav1.OwnerReference {
	rcvk := middlev1alpha1.GroupVersion.WithKind(rp.Kind)
	return []metav1.OwnerReference{
		*metav1.NewControllerRef(rp, rcvk),
	}
}

func (r *RedisProxyHandler) getLabels(rp *middlev1alpha1.RedisProxy) map[string]string {
	dynLabels := map[string]string{
		"redis/v1beta1": fmt.Sprintf("%s%c%s", rp.Namespace, '_', rp.Name),
	}
	defaultLabels := map[string]string{
		"redis/managed-by": "redis-proxy",
	}

	return util.MergeMap(defaultLabels, dynLabels, rp.Labels)
}
