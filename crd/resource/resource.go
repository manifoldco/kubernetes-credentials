package resource

import (
	"github.com/manifoldco/k8s-credentials/helpers/crd"
	apiextensionsclient "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// CreateCRD will create a new manifold.co Resource resource.
func CreateCRD(cs apiextensionsclient.Interface) error {
	return crd.CreateCRD(cs, "Resource", "resources", "manifold.co", "v1")
}

// Resource is the manifest representation of a manifold.co Resource CRD.
type Resource struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata"`
	Spec              Spec `json:"spec"`
}

// Spec is the specification that is required to build a valid Resource
// manifest.
type Spec struct {
	Resource    string           `json:"resource"`
	Team        string           `json:"string"`
	Credentials []CredentialSpec `json:"credentials,omitempty"`
}

// CredentialSpec is the specification that is required to filter out specific
// credentials in the Resource spec.
type CredentialSpec struct {
	Key     string `json:"key"`
	Name    string `json:"name"`
	Default string `json:"default"`
}
