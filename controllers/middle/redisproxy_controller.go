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

package middle

import (
	"context"
	"github.com/DevineLiu/redis-operator/controllers/middle/proxyservice"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	middlev1alpha1 "github.com/DevineLiu/redis-operator/apis/middle/v1alpha1"
	"github.com/DevineLiu/redis-operator/controllers/middle/client/k8s"
	"github.com/DevineLiu/redis-operator/controllers/middle/redisproxy"
	"github.com/go-logr/logr"
)

// RedisProxyReconciler reconciles a RedisProxy object
type RedisProxyReconciler struct {
	client.Client
	Scheme  *runtime.Scheme
	Logger  logr.Logger
	Record  record.EventRecorder
	Handler *redisproxy.RedisProxyHandler

}

//+kubebuilder:rbac:groups=middle.alauda.cn,resources=redisproxies,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=middle.alauda.cn,resources=redisproxies/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=middle.alauda.cn,resources=redisproxies/finalizers,verbs=update

func (r *RedisProxyReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	_ = log.FromContext(ctx)
	instance := &middlev1alpha1.RedisProxy{}
	err := r.Client.Get(ctx, req.NamespacedName, instance)
	if err != nil {
		if errors.IsNotFound(err) {
			return reconcile.Result{}, nil
		}
		return reconcile.Result{}, err
	}
	if err = r.Handler.Do(instance); err != nil {
		return reconcile.Result{}, err
	}
	return ctrl.Result{RequeueAfter: time.Duration(ReconcileTime) * time.Second}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *RedisProxyReconciler) SetupWithManager(mgr ctrl.Manager) error {
	r.SetupEventRecord(mgr)
	r.SetupHandler(mgr)
	return ctrl.NewControllerManagedBy(mgr).
		For(&middlev1alpha1.RedisProxy{}).
		Owns(&corev1.ConfigMap{}).
		Owns(&appsv1.Deployment{}).
		Owns(&corev1.Service{}).
		Complete(r)
}

func (r *RedisProxyReconciler) SetupEventRecord(mgr ctrl.Manager) {
	r.Record = mgr.GetEventRecorderFor("redis-proxy")
}

func (r *RedisProxyReconciler) SetupHandler(mgr ctrl.Manager) {
	k8sService := k8s.New(mgr.GetClient(), r.Logger)
	//redisClient := redis.New()
	rpkc := proxyservice.NewRedisProxyKubeClient(k8sService, r.Logger, r.Client.Status(), r.Record)

	r.Handler = &redisproxy.RedisProxyHandler{
		Logger:     r.Logger,
		Record:     r.Record,
		K8sService: k8sService,
		RpServices: rpkc,
	}
}
