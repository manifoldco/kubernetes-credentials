package primitives

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/manifoldco/go-manifold/integrations/primitives"
)

// Project is the manifest representation of a manifold.co Project CRD.
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type Project struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata"`
	Spec              *ProjectSpec `json:"spec"`
}

// ProjectList represents a list of available ProjectConfigurations in the
// cluster.
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type ProjectList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`
	Items           []*Project `json:"items"`
}

// ProjectSpec is the specification that is required to build a valid Project
// manifest.
type ProjectSpec struct {
	Name      string          `json:"project,name"`
	Team      string          `json:"team,omitempty"`
	Resources []*ResourceSpec `json:"resources,omitempty"`
}

// ManifoldPrimitive converts the ProjectSpec to a manifold project integration
// primitive.
func (ps *ProjectSpec) ManifoldPrimitive() *primitives.Project {
	resources := make([]*primitives.Resource, len(ps.Resources))
	for i, r := range ps.Resources {
		resources[i] = r.ManifoldPrimitive()
	}

	return &primitives.Project{
		Name:      ps.Name,
		Team:      ps.Team,
		Resources: resources,
	}
}
