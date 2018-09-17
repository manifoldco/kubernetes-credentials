package controller

import (
	"context"
	"encoding/base64"
	"fmt"
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

	"github.com/manifoldco/go-manifold/integrations"
	"github.com/manifoldco/kubernetes-credentials/crd"
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
	mc        *integrations.Client
	namespace string
}

// New returns a new controller
func New(kc *kubernetes.Clientset, rc *rest.RESTClient, mc *integrations.Client) *Controller {
	return &Controller{kc: kc, rc: rc, mc: mc}
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

	creds, err := c.mc.GetResourcesCredentialValues(ctx, &project.Spec.Name, project.Spec.ManifoldPrimitive().Resources)
	if err != nil {
		log.Printf("Error getting the credentials for %s: %s", project.Spec, err)
		return
	}

	cmap, err := integrations.FlattenResourcesCredentialValues(creds)
	if err != nil {
		log.Print("Error flattening credentials:", err)
		return
	}

	// determine if we need to decode values or not
	encodingKeys := map[string]string{}
	for _, resource := range project.Spec.Resources {
		encodingResourceKeys(resource, encodingKeys)
	}

	secretData := decodedByteMap(cmap, encodingKeys)
	c.createOrUpdateSecret(&project.ObjectMeta, secretData, project.Spec.SecretType(), projectControllerKind)
}

func (c *Controller) onProjectDelete(obj interface{}) {
	project := obj.(*primitives.Project)
	c.kc.Core().Secrets(project.Namespace).Delete(project.Name, &metav1.DeleteOptions{})
}

func (c *Controller) onResourceAdd(obj interface{})         { c.createOrUpdateResource(obj) }
func (c *Controller) onResourceUpdate(old, new interface{}) { c.createOrUpdateResource(new) }

func (c *Controller) createOrUpdateResource(obj interface{}) {
	resource := obj.(*primitives.Resource)
	ctx := context.Background()

	creds, err := c.mc.GetResourceCredentialValues(ctx, nil, resource.Spec.ManifoldPrimitive())
	if err != nil {
		log.Print("Error getting the credentials:", err)
		return
	}

	cmap, err := integrations.FlattenResourceCredentialValues(creds)
	if err != nil {
		log.Print("Error flattening credentials:", err)
		return
	}

	// determine if we need to decode values or not
	encodingKeys := map[string]string{}
	encodingResourceKeys(resource.Spec, encodingKeys)

	secretData := decodedByteMap(cmap, encodingKeys)
	c.createOrUpdateSecret(&resource.ObjectMeta, secretData, resource.Spec.SecretType(), resourceControllerKind)
}
func (c *Controller) onResourceDelete(obj interface{}) {
	resource := obj.(*primitives.Resource)
	c.kc.Core().Secrets(resource.Namespace).Delete(resource.Name, &metav1.DeleteOptions{})
}

func (c *Controller) createOrUpdateSecret(meta *metav1.ObjectMeta, secrets map[string][]byte, secretType v1.SecretType, gkv schema.GroupVersionKind) {
	data, err := secretData(secrets, secretType)
	if err != nil {
		log.Print("Error creating secret data: ", err)
		return
	}

	secret := v1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      meta.Name,
			Namespace: meta.Namespace,
			OwnerReferences: []metav1.OwnerReference{
				*metav1.NewControllerRef(meta, gkv),
			},
		},
		Data: data,
		Type: secretType,
	}

	s := c.kc.Core().Secrets(meta.Namespace)
	_, err = s.Update(&secret)
	if apierrors.IsNotFound(err) {
		_, err = s.Create(&secret)
	}

	if err != nil {
		log.Print("Error syncing secret:", err)
	}
}

func decodeValue(encoding, value string) ([]byte, error) {
	switch encoding {
	case "base64":
		return base64.StdEncoding.DecodeString(value)
	default:
		return nil, fmt.Errorf("Encoding '%s' not supported", encoding)
	}
}

func encodingResourceKeys(r *primitives.ResourceSpec, keys map[string]string) {
	for _, cred := range r.Credentials {
		if cred.Encoding != "" {
			k := cred.Key
			if cred.Name != "" {
				k = cred.Name
			}

			keys[k] = cred.Encoding
		}
	}
}

func decodedByteMap(cmap, encodingKeys map[string]string) map[string][]byte {
	secretData := make(map[string][]byte)
	for k, v := range cmap {
		var bts = []byte(v)

		if e, ok := encodingKeys[k]; ok {
			var err error
			bts, err = decodeValue(e, v)
			if err != nil {
				log.Printf("Error decoding value for key '%s': %s", k, err.Error())
			}
		}

		secretData[k] = bts
	}

	return secretData
}
