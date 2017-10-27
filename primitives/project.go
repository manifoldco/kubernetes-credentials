package primitives

import (
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Project is the manifest representation of a manifold.co Project CRD.
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type Project struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata"`
	Spec              *ProjectSpec `json:"spec"`
}

// Valid will validate the Project.
func (p *Project) Valid() bool {
	if p.Spec == nil {
		fmt.Println("no project spec")
		return false
	}

	return p.Spec.Valid()
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

// Valid will validate the ProjectSpec.
func (p *ProjectSpec) Valid() bool {
	if p.Name == "" {
		fmt.Println("no project spec name")
		return false
	}

	for _, r := range p.Resources {
		if !r.Valid() {
			return false
		}
	}

	return true
}
