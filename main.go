package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"

	"github.com/manifoldco/go-manifold"

	"github.com/manifoldco/kubernetes-credentials/controller"
	"github.com/manifoldco/kubernetes-credentials/crd"
	"github.com/manifoldco/kubernetes-credentials/crd/projects"
	"github.com/manifoldco/kubernetes-credentials/crd/resources"
	"github.com/manifoldco/kubernetes-credentials/helpers/client"
	"github.com/manifoldco/kubernetes-credentials/primitives"
)

func main() {
	log.Printf("Starting the controller...")

	ctx, cancelFunc := context.WithCancel(context.Background())
	defer cancelFunc()

	cfg, err := rest.InClusterConfig()
	if err != nil {
		log.Fatal(err)
	}

	kc, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		log.Fatal(err)
	}

	cs, err := clientset.NewForConfig(cfg)
	if err != nil {
		log.Fatal(err)
	}

	if err := crd.CreateCRD(cs, primitives.CRDProjectsName, primitives.CRDProjectsPlural, primitives.CRDGroup, primitives.CRDVersion); err != nil {
		log.Fatal(err)
	}
	if err := crd.CreateCRD(cs, primitives.CRDResourcesName, primitives.CRDResourcesPlural, primitives.CRDGroup, primitives.CRDVersion); err != nil {
		log.Fatal(err)
	}

	rc, err := newClient(cfg)
	if err != nil {
		log.Fatal(err)
	}

	ptr := func(str string) *string {
		return &str
	}

	manifoldClient := manifold.New(manifold.WithAPIToken(os.Getenv("MANIFOLD_API_TOKEN")))
	wrapper, err := client.New(manifoldClient, ptr(os.Getenv("MANIFOLD_TEAM")))
	if err != nil {
		log.Fatal(err)
	}

	ctrl := controller.New(kc, rc, wrapper)
	go ctrl.Run(ctx)

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Printf("Shutting down...")
}

func newClient(cfg *rest.Config) (*rest.RESTClient, error) {
	scheme := runtime.NewScheme()

	if err := projects.AddToScheme(scheme); err != nil {
		return nil, err
	}

	if err := resources.AddToScheme(scheme); err != nil {
		return nil, err
	}

	config := *cfg
	config.GroupVersion = &crd.SchemeGroupVersion
	config.APIPath = "/apis"
	config.ContentType = runtime.ContentTypeJSON
	config.NegotiatedSerializer = serializer.DirectCodecFactory{CodecFactory: serializer.NewCodecFactory(scheme)}

	return rest.RESTClientFor(&config)
}
