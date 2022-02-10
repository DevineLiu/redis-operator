package base

import (
	"context"
	"encoding/json"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/apiutil"
	"strings"
)

func NewK8sClient(cfg *rest.Config) (*kubernetes.Clientset, error) {
	clientSet, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		return nil, err
	}
	return clientSet, nil
}

type DynamicClient struct {
	dynamic.Interface
	*runtime.Scheme
}

func (dynamicClient *DynamicClient) List(ctx context.Context, list runtime.Object, opts ...client.ListOption) error {
	panic("implement me")
}

func (dynamicClient *DynamicClient) Create(ctx context.Context, obj runtime.Object, opts ...client.CreateOption) error {
	resourceInterface, err := dynamicClient.ResourceInterfaceFromObj(obj)
	if err != nil {
		return err
	}

	toCreateUnstructured, err := dynamicClient.convertToUnstructured(obj)
	if err != nil {
		return err
	}
	createdUnstructured, err := resourceInterface.Create(ctx, toCreateUnstructured, v1.CreateOptions{})
	if err != nil {
		return err
	}
	return fillRuntimeObject(createdUnstructured, obj)
}

func (dynamicClient *DynamicClient) ResourceInterfaceFromObj(obj runtime.Object) (dynamic.ResourceInterface, error) {
	namespace, err := meta.NewAccessor().Namespace(obj)
	if err != nil {
		return nil, err
	}
	name, err := meta.NewAccessor().Name(obj)
	if err != nil {
		return nil, err
	}
	return dynamicClient.ResourceInterface(obj, types.NamespacedName{
		Name:      name,
		Namespace: namespace,
	})
}

func (dynamicClient *DynamicClient) Delete(ctx context.Context, obj runtime.Object, opts ...client.DeleteOption) error {
	panic("implement me")
}

func (dynamicClient *DynamicClient) Update(ctx context.Context, obj runtime.Object, opts ...client.UpdateOption) error {
	resourceInterface, err := dynamicClient.ResourceInterfaceFromObj(obj)
	if err != nil {
		return err
	}

	toUpdateUnstructured, err := dynamicClient.convertToUnstructured(obj)
	if err != nil {
		return err
	}
	updatedUnstructured, err := resourceInterface.Update(ctx, toUpdateUnstructured, v1.UpdateOptions{})
	if err != nil {
		return err
	}
	return fillRuntimeObject(updatedUnstructured, obj)
}

func (dynamicClient *DynamicClient) Patch(ctx context.Context, obj runtime.Object, patch client.Patch, opts ...client.PatchOption) error {
	panic("implement me")
}

func (dynamicClient *DynamicClient) DeleteAllOf(ctx context.Context, obj runtime.Object, opts ...client.DeleteAllOfOption) error {
	panic("implement me")
}

func (dynamicClient *DynamicClient) Status() client.StatusWriter {
	panic("implement me")
}

func NewNoCacheDynamic(cfg *rest.Config, Scheme *runtime.Scheme) (*DynamicClient, error) {
	clientSet, err := dynamic.NewForConfig(cfg)
	if err != nil {
		return nil, err
	}

	return &DynamicClient{
		clientSet,
		Scheme,
	}, nil
}

func (dynamicClient *DynamicClient) Get(ctx context.Context, namespacedName types.NamespacedName, obj runtime.Object) error {
	resourceInterface, err := dynamicClient.ResourceInterface(obj, namespacedName)
	if err != nil {
		return err
	}

	unstructuredObject, err := resourceInterface.Get(ctx, namespacedName.Name, v1.GetOptions{})
	if err != nil {
		return err
	}

	return fillRuntimeObject(unstructuredObject, obj)
}

func (dynamicClient *DynamicClient) ResourceInterface(obj runtime.Object, namespacedName types.NamespacedName) (dynamic.ResourceInterface, error) {
	var resourceInterface dynamic.ResourceInterface
	gvk, err := apiutil.GVKForObject(obj, dynamicClient.Scheme)
	if err != nil {
		return resourceInterface, err
	}
	groupVersionResource := getGvr(gvk, obj)
	namespaceableResourceInterface := dynamicClient.Resource(groupVersionResource)
	if namespacedName.Namespace != "" {
		resourceInterface = namespaceableResourceInterface.Namespace(namespacedName.Namespace)
	} else {
		resourceInterface = namespaceableResourceInterface
	}
	return resourceInterface, err
}

func getGvr(gvk schema.GroupVersionKind, obj runtime.Object) schema.GroupVersionResource {
	if strings.HasSuffix(gvk.Kind, "List") && meta.IsListType(obj) {
		gvk.Kind = gvk.Kind[:len(gvk.Kind)-4]
	}
	groupVersionResource := schema.GroupVersionResource{
		Group:    gvk.Group,
		Version:  gvk.Version,
		Resource: strings.ToLower(gvk.Kind) + "s",
	}
	return groupVersionResource
}

func fillRuntimeObject(unstructured *unstructured.Unstructured, obj runtime.Object) (err error) {
	marshal, err := json.Marshal(unstructured)
	if err != nil {
		return err
	}
	return json.Unmarshal(marshal, obj)
}

func (dynamicClient *DynamicClient) convertToUnstructured(obj runtime.Object) (*unstructured.Unstructured, error) {
	toUnstructured, err := runtime.DefaultUnstructuredConverter.ToUnstructured(obj)
	if err != nil {
		return nil, err
	}
	gvk, err := apiutil.GVKForObject(obj, dynamicClient.Scheme)
	if err != nil {
		return nil, err
	}
	toUnstructured["kind"] = gvk.Kind
	toUnstructured["apiVersion"] = gvk.Group + "/" + gvk.Version

	return &unstructured.Unstructured{
		Object: toUnstructured,
	}, nil
}
