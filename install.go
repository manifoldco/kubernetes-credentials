package main

import (
	"flag"

	apiextensionsclient "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	"k8s.io/client-go/tools/clientcmd"

	"github.com/manifoldco/k8s-credentials/crd/project"
	"github.com/manifoldco/k8s-credentials/crd/resource"
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

	apiextensionsclientset, err := apiextensionsclient.NewForConfig(config)
	if err != nil {
		panic(err)
	}

	project.CreateCRD(apiextensionsclientset)
	resource.CreateCRD(apiextensionsclientset)
}
