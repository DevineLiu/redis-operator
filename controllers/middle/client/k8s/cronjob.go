package k8s

import (
	"context"

	"github.com/go-logr/logr"
	v1 "k8s.io/api/batch/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type CronJob interface {
	GetCronJob(namespace, name string) (*v1.CronJob, error)
	ListCronJobs(namespace string, cl client.ListOptions) (*v1.CronJobList, error)
	DeleteCronJob(namespace, name string) error
	CreateCronJob(namespace string, cronjob *v1.CronJob) error
}

type CronJobOption struct {
	client client.Client
	logger logr.Logger
}

func NewCronJob(kubeClient client.Client, logger logr.Logger) CronJob {
	logger = logger.WithValues("service", "k8s.CronJob")
	return &CronJobOption{
		client: kubeClient,
		logger: logger,
	}
}

func (c *CronJobOption) GetCronJob(namespace, name string) (*v1.CronJob, error) {
	cronjob := &v1.CronJob{}
	err := c.client.Get(context.TODO(), types.NamespacedName{
		Name:      name,
		Namespace: namespace,
	}, cronjob)
	if err != nil {
		return nil, err
	}
	return cronjob, err
}

func (c *CronJobOption) ListCronJobs(namespace string, cl client.ListOptions) (*v1.CronJobList, error) {
	cs := &v1.CronJobList{}
	listOps := &cl
	err := c.client.List(context.TODO(), cs, listOps)
	return cs, err
}

func (c *CronJobOption) DeleteCronJob(namespace, name string) error {
	cronjob := &v1.CronJob{}
	if err := c.client.Get(context.TODO(), types.NamespacedName{
		Name:      name,
		Namespace: namespace,
	}, cronjob); err != nil {
		return err
	}
	return c.client.Delete(context.TODO(), cronjob)
}

func (c *CronJobOption) CreateCronJob(namespace string, cronjob *v1.CronJob) error {
	err := c.client.Create(context.TODO(), cronjob)
	if err != nil {
		return err
	}
	c.logger.WithValues("namespace", namespace, "cronjob", cronjob.ObjectMeta.Name).Info("cronjob created")
	return err
}
