package main

import (
	"context"
	"log"
	"time"

	"k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"

	"github.com/manifoldco/kubernetes-credentials/crd"
	"github.com/manifoldco/kubernetes-credentials/helpers/client"
	"github.com/manifoldco/kubernetes-credentials/primitives"
)

var (
	projectControllerKind  = crd.SchemeGroupVersion.WithKind("Project")
	resourceControllerKind = crd.SchemeGroupVersion.WithKind("Resource")
)

// Controller is the kubernetes controller that handles syncing Manifold
// credentials into kubernetes secrets.
type Controller struct {
	kc        *kubernetes.Clientset
	rc        *rest.RESTClient
	mc        *client.Client
	namespace string
}

// Run runs this controller
func (c *Controller) Run(ctx context.Context) error {
	if err := c.watchProjects(ctx); err != nil {
		log.Println("Failed to register project watcher:", err)
		return err
	}

	if err := c.watchResources(ctx); err != nil {
		log.Println("Failed to register resource watcher:", err)
		return err
	}

	return nil
}

// watch configures and runs a watcher for the given resource type. We use this
// to listen for changes on both projects and resources.
func (c *Controller) watch(ctx context.Context, resource string, obj runtime.Object, handler cache.ResourceEventHandler) error {
	source := cache.NewListWatchFromClient(c.rc, resource, c.namespace, fields.Everything())

	resyncPeriod := 10 * time.Second
	_, controller := cache.NewInformer(source, obj, resyncPeriod, handler)

	go controller.Run(ctx.Done())
	return nil

}

func (c *Controller) watchProjects(ctx context.Context) error {
	return c.watch(ctx, primitives.CRDProjectsPlural, &primitives.Project{}, cache.ResourceEventHandlerFuncs{
		AddFunc:    c.onProjectAdd,
		UpdateFunc: c.onProjectUpdate,
		DeleteFunc: c.onProjectDelete,
	})
}

func (c *Controller) watchResources(ctx context.Context) error {
	return c.watch(ctx, primitives.CRDResourcesPlural, &primitives.Resource{}, cache.ResourceEventHandlerFuncs{
		AddFunc:    c.onResourceAdd,
		UpdateFunc: c.onResourceUpdate,
		DeleteFunc: c.onResourceDelete,
	})
}

func (c *Controller) onProjectAdd(obj interface{})         { c.createOrUpdateProject(obj) }
func (c *Controller) onProjectUpdate(old, new interface{}) { c.createOrUpdateProject(new) }

func (c *Controller) createOrUpdateProject(obj interface{}) {
	project := obj.(*primitives.Project)
	ctx := context.Background()

	creds, err := c.mc.GetResourcesCredentialValues(ctx, &project.Spec.Name, project.Spec.Resources)
	if err != nil {
		log.Print("Error getting the credentials:", err)
		return
	}

	cmap, err := client.FlattenResourcesCredentialValues(creds)
	if err != nil {
		log.Print("Error flattening credentials:", err)
		return
	}

	secretData := make(map[string][]byte)
	for k, v := range cmap {
		secretData[k] = []byte(v)
	}

	c.createOrUpdateSecret(&project.ObjectMeta, secretData, projectControllerKind)
}

func (c *Controller) onProjectDelete(obj interface{}) {
	project := obj.(*primitives.Project)
	c.kc.Core().Secrets(project.Namespace).Delete(project.Name+"-credentials", &metav1.DeleteOptions{})
}

func (c *Controller) onResourceAdd(obj interface{})         { c.createOrUpdateResource(obj) }
func (c *Controller) onResourceUpdate(old, new interface{}) { c.createOrUpdateResource(new) }

func (c *Controller) createOrUpdateResource(obj interface{}) {
	resource := obj.(*primitives.Resource)
	ctx := context.Background()

	creds, err := c.mc.GetResourceCredentialValues(ctx, nil, resource.Spec)
	if err != nil {
		log.Print("Error getting the credentials:", err)
		return
	}

	cmap, err := client.FlattenResourceCredentialValues(creds)
	if err != nil {
		log.Print("Error flattening credentials:", err)
		return
	}

	secretData := make(map[string][]byte)
	for k, v := range cmap {
		secretData[k] = []byte(v)
	}

	c.createOrUpdateSecret(&resource.ObjectMeta, secretData, resourceControllerKind)
}
func (c *Controller) onResourceDelete(obj interface{}) {
	resource := obj.(*primitives.Resource)
	c.kc.Core().Secrets(resource.Namespace).Delete(resource.Name+"-credentials", &metav1.DeleteOptions{})
}

func (c *Controller) createOrUpdateSecret(meta *metav1.ObjectMeta, secrets map[string][]byte, gkv schema.GroupVersionKind) {
	secret := v1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      meta.Name + "-credentials",
			Namespace: meta.Namespace,
			OwnerReferences: []metav1.OwnerReference{
				*metav1.NewControllerRef(meta, gkv),
			},
		},
		Data: secrets,
	}

	s := c.kc.Core().Secrets(meta.Namespace)
	_, err := s.Update(&secret)
	if apierrors.IsNotFound(err) {
		_, err = s.Create(&secret)
	}

	if err != nil {
		log.Print("Error syncing secret:", err)
	}
}
