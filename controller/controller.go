package main

import (
	"context"
	"log"
	"time"

	"k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"

	"github.com/manifoldco/kubernetes-credentials/crd"
	"github.com/manifoldco/kubernetes-credentials/helpers/client"
	"github.com/manifoldco/kubernetes-credentials/primitives"
)

var (
	projectControllerKind = crd.SchemeGroupVersion.WithKind("Project")
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
		log.Printf("Failed to register project watcher: %s", err)
		return err
	}

	return nil
}

func (c *Controller) watchProjects(ctx context.Context) error {
	source := cache.NewListWatchFromClient(
		c.rc,
		primitives.CRDProjectsPlural,
		c.namespace,
		fields.Everything(),
	)

	_, controller := cache.NewInformer(
		source,

		// The object type.
		&primitives.Project{},

		// resyncPeriod
		// Every resyncPeriod, all resources in the cache will retrigger events.
		// Set to 0 to disable the resync.
		10*time.Second,

		// Your custom resource event handlers.
		cache.ResourceEventHandlerFuncs{
			AddFunc:    c.onProjectAdd,
			UpdateFunc: c.onProjectUpdate,
			DeleteFunc: c.onProjectDelete,
		},
	)

	go controller.Run(ctx.Done())
	return nil
}

func (c *Controller) onProjectAdd(obj interface{}) { c.createOrUpdateSecret(obj) }

func (c *Controller) onProjectUpdate(old, new interface{}) {
	c.createOrUpdateSecret(new)
}

func (c *Controller) onProjectDelete(obj interface{}) {
	log.Printf("Received a delete for a project crd")

	project := obj.(*primitives.Project)
	c.kc.Core().Secrets(project.Namespace).Delete(project.Name+"-credentials", &metav1.DeleteOptions{})
}

func (c *Controller) createOrUpdateSecret(obj interface{}) {
	project := obj.(*primitives.Project)
	ctx := context.Background()

	creds, err := c.mc.GetResourcesCredentialValues(ctx, &project.Spec.Label, project.Spec.Resources)
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

	secret := v1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      project.Name + "-credentials",
			Namespace: project.Namespace,
			OwnerReferences: []metav1.OwnerReference{
				*metav1.NewControllerRef(project, projectControllerKind),
			},
		},
		Data: secretData,
	}

	s := c.kc.Core().Secrets(project.Namespace)
	_, err = s.Update(&secret)
	if apierrors.IsNotFound(err) {
		_, err = s.Create(&secret)
	}

	if err != nil {
		log.Print("Error syncing secret:", err)
	}
}
