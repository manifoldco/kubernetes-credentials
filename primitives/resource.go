package primitives

import (
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Resource is the manifest representation of a manifold.co Resource CRD.
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type Resource struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata"`
	Spec              *ResourceSpec `json:"spec"`
}

// Valid will validate the resource.
func (r *Resource) Valid() bool {
	if r.Spec == nil {
		fmt.Println("no resource spec")
		return false
	}

	return r.Spec.Valid()
}

// ResourceList represents a list of available ResourceConfigurations in the
// cluster.
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type ResourceList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`
	Items           []*Resource `json:"items"`
}

// Spec is the specification that is required to build a valid Resource
// manifest.
type ResourceSpec struct {
	Label       string            `json:"resource,label"`
	Team        string            `json:"team,omitempty"`
	Credentials []*CredentialSpec `json:"credentials,omitempty"`
}

// Valid will validate the ResourceSpec.
func (r *ResourceSpec) Valid() bool {
	if r.Label == "" {
		fmt.Println("no resource spec label")
		return false
	}

	for _, c := range r.Credentials {
		if !c.Valid() {
			return false
		}
	}

	return true
}
