package project

import (
	"github.com/manifoldco/k8s-credentials/crd/resource"
	"github.com/manifoldco/k8s-credentials/helpers/crd"
	apiextensionsclient "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// CreateCRD will create a new manifold.co Project resource.
func CreateCRD(cs apiextensionsclient.Interface) error {
	return crd.CreateCRD(cs, "Project", "projects", "manifold.co", "v1")
}

// Project is the manifest representation of a manifold.co Project CRD.
type Project struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata"`
	Spec              Spec `json:"spec"`
}

// Spec is the specification that is required to build a valid Project manifest.
type Spec struct {
	Project   string          `json:"project"`
	Team      string          `json:"team"`
	Resources []resource.Spec `json:"resources,omitempty"`
}
