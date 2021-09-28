package redisfailover

import (
	"context"
	"fmt"
	middlev1alpha1 "github.com/DevineLiu/redis-operator/apis/middle/v1alpha1"
	"github.com/DevineLiu/redis-operator/controllers/middle/client/k8s"
	"github.com/DevineLiu/redis-operator/controllers/middle/service"
	util "github.com/DevineLiu/redis-operator/controllers/util"
	"github.com/go-logr/logr"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type RedisFailoverHandler struct {
	Logger       logr.Logger
	Record       record.EventRecorder
	K8sService   k8s.Service
	RfServices   service.RedisFailoverClient
	RfChecker    service.RedisFailoverCheck
	RfHealer     service.RedisFailoverHeal
	StatusWriter client.Client
}

func (r *RedisFailoverHandler) Do(rf *middlev1alpha1.RedisFailover) error {
	if err := rf.Validate(); err != nil {
		r.Record.Event(rf, v1.EventTypeWarning, "Valiadte", fmt.Sprintf("err: %s", err.Error()))
		return err
	}
	oRefs := r.createOwnerReferences(rf)
	labels := r.getLabels(rf)

	r.Logger.WithValues("namespace", rf.Namespace, "name", rf.Name).V(2).Info("Ensure...")
	r.Record.Event(rf, v1.EventTypeNormal, "Ensure", "Ensure running")

	if err := r.Ensure(rf, labels, oRefs); err != nil {
		r.Record.Event(rf, v1.EventTypeWarning, "EnsureError", err.Error())
		rf.Status.SetFailedCondition(err.Error())
		r.StatusWriter.Status().Update(context.TODO(), rf)
		return err
	}

	r.Logger.WithValues("namespace", rf.Namespace, "name", rf.Name).V(2).Info("CheckAndHeal...")
	r.Record.Event(rf, v1.EventTypeNormal, "Heal", "CheckAndHeal")
	if err := r.CheckAndHeal(rf); err != nil {
		r.Logger.WithValues("namespace", rf.Namespace, "name", rf.Name).V(2).Info("CheckAndHealError: %s", err.Error())
		if rf.Status.IsLastConditionWaitingPodReady() {
			r.Record.Event(rf, v1.EventTypeNormal, "CreateCluster", "CreateCluster for waiting pod ")
			r.StatusWriter.Status().Update(context.TODO(), rf)
		} else {
			r.Record.Event(rf, v1.EventTypeWarning, "CheckAndHealError", err.Error())
			rf.Status.SetFailedCondition(err.Error())
			r.StatusWriter.Status().Update(context.TODO(), rf)
			return err
		}
		return err
	}
	r.Logger.WithValues("namespace", rf.Namespace, "name", rf.Name).V(2).Info("SetReadyCondition...")
	r.Record.Event(rf, v1.EventTypeNormal, "HEALTH", "Cluster Be Healthly")
	rf.Status.SetReadyCondition("HEALTHLY")
	r.StatusWriter.Status().Update(context.TODO(), rf)
	return nil
}

func (r *RedisFailoverHandler) getLabels(rf *middlev1alpha1.RedisFailover) map[string]string {
	dynLabels := map[string]string{
		"redis/v1beta1": fmt.Sprintf("%s%c%s", rf.Namespace, '_', rf.Name),
	}
	defaultLabels := map[string]string{
		"redis/managed-by": "redis-operator",
	}

	return util.MergeMap(defaultLabels, dynLabels, rf.Labels)
}

func (r *RedisFailoverHandler) createOwnerReferences(rf *middlev1alpha1.RedisFailover) []metav1.OwnerReference {
	rcvk := middlev1alpha1.GroupVersion.WithKind(rf.Kind)
	return []metav1.OwnerReference{
		*metav1.NewControllerRef(rf, rcvk),
	}
}
