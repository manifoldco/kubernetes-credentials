package main

import (
	"context"
	"fmt"
	"log"

	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"

	"github.com/manifoldco/kubernetes-credentials/helpers/client"
	"github.com/manifoldco/kubernetes-credentials/primitives"
)

type Controller struct {
	cClient   *rest.RESTClient
	mc        *client.Client
	namespace string
}

func (c *Controller) Run(ctx context.Context) error {
	if err := c.watchProjects(ctx); err != nil {
		log.Printf("Failed to register project watcher: %s", err)
		return err
	}

	return nil
}

func (c *Controller) watchProjects(ctx context.Context) error {
	source := cache.NewListWatchFromClient(
		c.cClient,
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
		0,

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

func (c *Controller) onProjectAdd(obj interface{}) {
	project := obj.(*primitives.Project)
	fmt.Println(project.Spec)
	ctx := context.Background()

	label := primitives.NewLabel(project.Spec.Label)
	creds, err := c.mc.GetResourcesCredentialValues(ctx, label, project.Spec.Resources)
	if err != nil {
		log.Printf("Error getting the credentials: %s", err.Error())
		return
	}

	fmt.Println(creds)
}

func (c *Controller) onProjectUpdate(old, new interface{}) {
	log.Printf("Received an update for a CRD!")
}

func (c *Controller) onProjectDelete(obj interface{}) {
	log.Printf("Received a delete for a CRD!")
}
