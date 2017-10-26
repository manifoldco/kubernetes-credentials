package projects

import (
	"k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"

	"github.com/manifoldco/kubernetes-credentials/crd"
	"github.com/manifoldco/kubernetes-credentials/primitives"
)

// The project scheme builder
var (
	SchemeBuilder = runtime.NewSchemeBuilder(addKnownTypes)
	AddToScheme   = SchemeBuilder.AddToScheme
)

// addKnownTypes adds the set of types defined in this package to the supplied scheme.
func addKnownTypes(scheme *runtime.Scheme) error {
	scheme.AddKnownTypes(crd.SchemeGroupVersion,
		&primitives.Project{},
		&primitives.ProjectList{},
	)
	v1.AddToGroupVersion(scheme, crd.SchemeGroupVersion)
	return nil
}
