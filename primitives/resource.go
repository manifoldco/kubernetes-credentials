package primitives

import (
	"github.com/manifoldco/go-manifold/integrations/primitives"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Resource is the manifest representation of a manifold.co Resource CRD.
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type Resource struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata"`
	Spec              *ResourceSpec `json:"spec"`
}

// ResourceList represents a list of available ResourceConfigurations in the
// cluster.
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type ResourceList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`
	Items           []*Resource `json:"items"`
}

// ResourceSpec is the specification that is required to build a valid Resource
// manifest.
type ResourceSpec struct {
	Name        string            `json:"resource,name"`
	Team        string            `json:"team,omitempty"`
	Credentials []*CredentialSpec `json:"credentials,omitempty"`
}

// ManifoldPrimitive converts the ResourceSpec to a manifold project integration
// primitive.
func (rs *ResourceSpec) ManifoldPrimitive() *primitives.Resource {
	credentials := make([]*primitives.Credential, len(rs.Credentials))
	for i, c := range rs.Credentials {
		credentials[i] = c.ManifoldPrimitive()
	}

	return &primitives.Resource{
		Name:        rs.Name,
		Team:        rs.Team,
		Credentials: credentials,
	}
}
