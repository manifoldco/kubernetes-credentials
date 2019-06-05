package controller

import (
	"context"
	"encoding/base64"
	"fmt"
	"time"

	log "github.com/sirupsen/logrus"
	v1 "k8s.io/api/core/v1"
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
		log.WithError(err).Error("could not register project watcher")
		return err
	}

	if err := c.watchResources(ctx); err != nil {
		log.WithError(err).Error("could not register resource watcher")
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
	l := log.WithFields(log.Fields{
		"crd_name":      project.Name,
		"crd_namespace": project.Namespace,
		"project":       project.Spec.Name,
		"team":          project.Spec.Team,
		"type":          project.Spec.Type,
	})

	creds, err := c.mc.GetResourcesCredentialValues(ctx, &project.Spec.Name, project.Spec.ManifoldPrimitive().Resources)
	if err != nil {
		l.WithError(err).Error("could not get project credentials")
		return
	}

	cmap, err := integrations.FlattenResourcesCredentialValues(creds)
	if err != nil {
		l.WithError(err).Error("could not flatten project credentials")
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
	if err := c.kc.Core().Secrets(project.Namespace).Delete(project.Name, &metav1.DeleteOptions{}); err != nil {
		log.WithError(err).Error("issue deleting the project")
	}
}

func (c *Controller) onResourceAdd(obj interface{})         { c.createOrUpdateResource(obj) }
func (c *Controller) onResourceUpdate(old, new interface{}) { c.createOrUpdateResource(new) }

func (c *Controller) createOrUpdateResource(obj interface{}) {
	resource := obj.(*primitives.Resource)
	ctx := context.Background()
	l := log.WithFields(log.Fields{
		"crd_name":      resource.Name,
		"crd_namespace": resource.Namespace,
		"resource":      resource.Spec.Name,
		"team":          resource.Spec.Team,
		"type":          resource.Spec.Type,
	})

	creds, err := c.mc.GetResourceCredentialValues(ctx, nil, resource.Spec.ManifoldPrimitive())
	if err != nil {
		l.WithError(err).Error("could not get resource credentials")
		return
	}

	cmap, err := integrations.FlattenResourceCredentialValues(creds)
	if err != nil {
		l.WithError(err).Error("could not flatten resource credentials")
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
	if err := c.kc.Core().Secrets(resource.Namespace).Delete(resource.Name, &metav1.DeleteOptions{}); err != nil {
		log.WithError(err).Error("issue deleting the resource")
	}
}

func (c *Controller) createOrUpdateSecret(meta *metav1.ObjectMeta, secrets map[string][]byte, secretType v1.SecretType, gkv schema.GroupVersionKind) {
	l := log.WithFields(log.Fields{
		"crd_name":      meta.Name,
		"crd_namespace": meta.Namespace,
	})

	data, err := secretData(secrets, secretType)
	if err != nil {
		l.WithError(err).Error("could not create secret")
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
		l.WithError(err).Error("could not sync secret")
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
				log.WithField("key", k).WithError(err).Error("could not decode value")
			}
		}

		secretData[k] = bts
	}

	return secretData
}
