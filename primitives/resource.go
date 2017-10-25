package primitives

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Resource is the manifest representation of a manifold.co Resource CRD.
type Resource struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata"`
	Spec              *ResourceSpec `json:"spec"`
}

// Valid will validate the Resource.
func (r *Resource) Valid() bool {
	if r.Spec == nil || r.Spec.Label == "" {
		return false
	}

	for _, c := range r.Spec.Credentials {
		if !c.Valid() {
			return false
		}
	}

	return true
}

// Spec is the specification that is required to build a valid Resource
// manifest.
type ResourceSpec struct {
	Label       string            `json:"resource,label"`
	Team        string            `json:"team,omitempty"`
	Credentials []*CredentialSpec `json:"credentials,omitempty"`
}
