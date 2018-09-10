package primitives

import (
	"strings"

	"github.com/manifoldco/go-manifold/integrations/primitives"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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
	Type      string          `json:"type,omitempty"`
	Resources []*ResourceSpec `json:"resources,omitempty"`
}

// SecretType returns the type of secret that should be generated for this spec.
func (ps *ProjectSpec) SecretType() v1.SecretType {
	t, err := secretType(ps.Type)

	// TODO jelmer: once we've put in validation, we can ignore this error.
	if err != nil {
		panic(err)
	}

	return t
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

// String returns a string representation of the spec
func (ps *ProjectSpec) String() string {
	var result []string

	if ps.Name != "" {
		result = append(result, "project: "+ps.Name)
	}

	if ps.Team != "" {
		result = append(result, "team: "+ps.Team)
	}

	if ps.Type != "" {
		result = append(result, "type: "+ps.Type)
	}

	return strings.Join(result, ", ")
}
