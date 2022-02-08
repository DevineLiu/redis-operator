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

package databases

import (
	"context"
	"fmt"
	"time"

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	databasesv1 "github.com/DevineLiu/redis-operator/apis/databases/v1"
	"github.com/DevineLiu/redis-operator/controllers/databases/redisfailover"
	"github.com/DevineLiu/redis-operator/controllers/databases/service"
	"github.com/DevineLiu/redis-operator/controllers/middle/client/k8s"
	"github.com/DevineLiu/redis-operator/controllers/middle/client/redis"
	"github.com/go-logr/logr"
)

const ReconcileTime = 60

// RedisFailoverReconciler reconciles a RedisFailover object
type RedisFailoverReconciler struct {
	Client  client.Client
	Scheme  *runtime.Scheme
	Logger  logr.Logger
	Record  record.EventRecorder
	Handler *redisfailover.RedisFailoverHandler
}

//+kubebuilder:rbac:groups=databases.spotahome.com,resources=redisfailovers,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=databases.spotahome.com,resources=redisfailovers/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=databases.spotahome.com,resources=redisfailovers/finalizers,verbs=update

func (r *RedisFailoverReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	instance := &databasesv1.RedisFailover{}
	err := r.Client.Get(ctx, req.NamespacedName, instance)
	if err != nil {
		if errors.IsNotFound(err) {
			return reconcile.Result{}, nil
		}
		return reconcile.Result{}, err
	}

	if err = r.Handler.Do(instance); err != nil {
		if instance.Status.IsWaitingPodReady() {
			r.Logger.WithValues("namespace", instance.Namespace, "name", instance.Name).V(2).Info("waiting pod ready", err.Error())
			return reconcile.Result{RequeueAfter: 20 * time.Second}, nil
		}
		return reconcile.Result{}, err
	}

	if err = r.Handler.RfChecker.CheckSentinelReadyReplicas(instance); err != nil {

		r.Logger.Info(err.Error())
		return reconcile.Result{RequeueAfter: 20 * time.Second}, nil
	}

	r.Logger.V(5).Info(fmt.Sprintf("RedisFailover Spec:\n %+v", instance))

	return ctrl.Result{RequeueAfter: time.Duration(ReconcileTime) * time.Second}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *RedisFailoverReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&databasesv1.RedisFailover{}).
		Complete(r)
}

// SetupEventRecord setup event  record for controller
func (r *RedisFailoverReconciler) SetupEventRecord(mgr ctrl.Manager) {
	r.Record = mgr.GetEventRecorderFor("redis-operator")
}

func (r *RedisFailoverReconciler) SetupHandler(mgr ctrl.Manager) {
	k8sService := k8s.New(mgr.GetClient(), r.Logger)
	redisClient := redis.New()
	rfkc := service.NewRedisFailoverKubeClient(k8sService, r.Logger, r.Client.Status(), r.Record)
	rfchecker := service.NewRedisFailoverChecker(k8sService, r.Logger, r.Client.Status(), r.Record, redisClient)
	rfhealer := service.NewRedisFailoverHealer(k8sService, r.Logger, r.Client.Status(), r.Record, redisClient)
	status := redisfailover.StatusWriter{
		Client: r.Client,
		Ctx:    context.TODO(),
	}

	r.Handler = &redisfailover.RedisFailoverHandler{
		Logger:       r.Logger,
		Record:       r.Record,
		K8sService:   k8sService,
		RfServices:   rfkc,
		RfChecker:    rfchecker,
		RfHealer:     rfhealer,
		StatusWriter: status,
	}

}
