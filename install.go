package main

import (
	"flag"

	apiextensionsclient "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	"k8s.io/client-go/tools/clientcmd"

	"github.com/manifoldco/k8s-credentials/helpers/crd"
)

func main() {
	masterURL := flag.String("master", "", "The address of the Kubernetes API server. Overrides any value in kubeconfig. Only required if out-of-cluster.")
	kubeconfig := flag.String("kubeconfig", "", "Path to a kube config. Only required if out-of-cluster.")
	flag.Parse()

	// Create the client config. Use masterURL and kubeconfig if given, otherwise assume in-cluster.
	config, err := clientcmd.BuildConfigFromFlags(*masterURL, *kubeconfig)
	if err != nil {
		panic(err)
	}

	cs, err := apiextensionsclient.NewForConfig(config)
	if err != nil {
		panic(err)
	}

	if err := crd.CreateCRD(cs, "Project", "projects", "manifold.co", "v1"); err != nil {
		panic(err)
	}
	if err := crd.CreateCRD(cs, "Resource", "resources", "manifold.co", "v1"); err != nil {
		panic(err)
	}
}
