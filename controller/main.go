package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/client-go/rest"

	manifold "github.com/manifoldco/go-manifold"
	"github.com/manifoldco/kubernetes-credentials/crd"
	"github.com/manifoldco/kubernetes-credentials/crd/projects"
	"github.com/manifoldco/kubernetes-credentials/crd/resources"
	"github.com/manifoldco/kubernetes-credentials/helpers/client"
)

func main() {
	log.Printf("Starting the controller...")

	ctx, cancelFunc := context.WithCancel(context.Background())
	defer cancelFunc()

	cfg, err := rest.InClusterConfig()
	if err != nil {
		log.Fatal(err)
	}

	restClient, _, err := newClient(cfg)
	if err != nil {
		log.Fatal(err)
	}

	ptr := func(str string) *string {
		return &str
	}

	fmt.Println(os.Getenv("MANIFOLD_API_TOKEN"), os.Getenv("MANIFOLD_TEAM"))
	manifoldClient := manifold.New(manifold.WithAPIToken(os.Getenv("MANIFOLD_API_TOKEN")))
	wrapper, err := client.New(manifoldClient, ptr(os.Getenv("MANIFOLD_TEAM")))
	if err != nil {
		log.Fatal(err)
	}

	ctrl := &Controller{
		mc:      wrapper,
		cClient: restClient,
	}
	go ctrl.Run(ctx)

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Printf("Shutting down...")
}

func newClient(cfg *rest.Config) (*rest.RESTClient, *runtime.Scheme, error) {
	scheme := runtime.NewScheme()

	if err := projects.AddToScheme(scheme); err != nil {
		return nil, nil, err
	}

	if err := resources.AddToScheme(scheme); err != nil {
		return nil, nil, err
	}

	config := *cfg
	config.GroupVersion = &crd.SchemeGroupVersion
	config.APIPath = "/apis"
	config.ContentType = runtime.ContentTypeJSON
	config.NegotiatedSerializer = serializer.DirectCodecFactory{CodecFactory: serializer.NewCodecFactory(scheme)}

	client, err := rest.RESTClientFor(&config)
	if err != nil {
		return nil, nil, err
	}

	return client, scheme, nil
}
