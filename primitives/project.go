package primitives

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Project is the manifest representation of a manifold.co Project CRD.
type Project struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata"`
	Spec              *ProjectSpec `json:"spec"`
}

// Valid will validate the Project.
func (p *Project) Valid() bool {
	if p.Spec == nil || p.Spec.Label == "" {
		return false
	}

	for _, r := range p.Spec.Resources {
		if !r.Valid() {
			return false
		}
	}

	return true
}

// ProjectSpec is the specification that is required to build a valid Project
// manifest.
type ProjectSpec struct {
	Label     string      `json:"project,label"`
	Team      string      `json:"team,omitempty"`
	Resources []*Resource `json:"resources,omitempty"`
}
