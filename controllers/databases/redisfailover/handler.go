package redisfailover

import (
	"context"
	"fmt"
	"regexp"

	databasesv1 "github.com/DevineLiu/redis-operator/apis/databases/v1"
	"github.com/DevineLiu/redis-operator/controllers/databases/service"
	"github.com/DevineLiu/redis-operator/controllers/middle/client/k8s"
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
	K8sService   k8s.Services
	RfServices   service.RedisFailoverClient
	RfChecker    service.RedisFailoverCheck
	RfHealer     service.RedisFailoverHeal
	StatusWriter StatusWriter
}

func (r *RedisFailoverHandler) Do(rf *databasesv1.RedisFailover) error {
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
		rf.Status.SetFailedPhase(err.Error())
		r.StatusWriter.Update(rf)
		return err
	}

	r.Logger.WithValues("namespace", rf.Namespace, "name", rf.Name).V(2).Info("CheckAndHeal...")
	r.Record.Event(rf, v1.EventTypeNormal, "Heal", "CheckAndHeal")
	if err := r.CheckAndHeal(rf); err != nil {
		r.Logger.WithValues("namespace", rf.Namespace, "name", rf.Name).V(2).Info("CheckAndHealError: %s", err.Error())
		if rf.Status.IsWaitingPodReady() {
			r.Record.Event(rf, v1.EventTypeNormal, "CreateCluster", err.Error())
		} else {
			r.Record.Event(rf, v1.EventTypeWarning, "CheckAndHealError", err.Error())
			rf.Status.SetFailedPhase(err.Error())
			r.StatusWriter.Update(rf)
			return err
		}
		return err
	}
	r.Logger.WithValues("namespace", rf.Namespace, "name", rf.Name).V(2).Info("SetReadyCondition...")
	r.Record.Event(rf, v1.EventTypeNormal, "HEALTH", "Cluster Be Healthly")
	rf.Status.SetReady("HEALTHLY")
	r.StatusWriter.Update(rf)
	return nil
}

const (
	rfLabelManagedByKey = "app.kubernetes.io/managed-by"
	rfLabelNameKey      = "redisfailovers.databases.spotahome.com/name"
)

func (r *RedisFailoverHandler) getLabels(rf *databasesv1.RedisFailover) map[string]string {
	dynLabels := map[string]string{
		rfLabelNameKey: rf.Name,
	}
	filteredCustomLabels := make(map[string]string)
	if rf.Spec.LabelWhitelist != nil && len(rf.Spec.LabelWhitelist) != 0 {
		for _, regex := range rf.Spec.LabelWhitelist {
			compiledRegexp, err := regexp.Compile(regex)
			if err != nil {
				r.Logger.Error(err, "Unable to compile label whitelist regex '%s', ignoring it.", regex)
				continue
			}
			for labelKey, labelValue := range rf.Labels {
				if match := compiledRegexp.MatchString(labelKey); match {
					filteredCustomLabels[labelKey] = labelValue
				}
			}
		}
	} else {
		// If no whitelist is specified then don't filter the labels.
		filteredCustomLabels = rf.Labels
	}
	defaultLabels := map[string]string{
		rfLabelManagedByKey: "redis-operator",
	}

	return util.MergeMap(defaultLabels, dynLabels, filteredCustomLabels)
}

func (r *RedisFailoverHandler) createOwnerReferences(rf *databasesv1.RedisFailover) []metav1.OwnerReference {
	rcvk := databasesv1.GroupVersion.WithKind(rf.Kind)
	return []metav1.OwnerReference{
		*metav1.NewControllerRef(rf, rcvk),
	}
}

type StatusWriter struct {
	client.Client
	Ctx context.Context
}

type StatusWrite interface {
	Update(rf *databasesv1.RedisFailover, opts ...client.UpdateOption) error
	Patch(rf *databasesv1.RedisFailover, patch client.Patch, opts ...client.PatchOption) error
}

func (s *StatusWriter) Patch(ctx context.Context, rf *databasesv1.RedisFailover, patch client.Patch, opts ...client.PatchOption) error {
	err := s.Status().Patch(s.Ctx, rf, patch, opts...)
	return err
}

func (s *StatusWriter) Update(rf *databasesv1.RedisFailover, opts ...client.UpdateOption) error {
	err := s.Status().Update(s.Ctx, rf, opts...)
	return err
}
