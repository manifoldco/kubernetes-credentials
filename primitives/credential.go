package primitives

import "github.com/manifoldco/go-manifold/integrations/primitives"

// CredentialSpec represents the specification that is required to filter out
// specific credentials in the Resource spec.
type CredentialSpec struct {
	Key      string `json:"key"`
	Name     string `json:"name,omitempty"`
	Default  string `json:"default,omitempty"`
	Encoding string `json:"encoding,omitempty"`
}

// ManifoldPrimitive converts the CredentialSpec to a manifold project integration
// primitive.
func (cs *CredentialSpec) ManifoldPrimitive() *primitives.Credential {
	return &primitives.Credential{
		Key:     cs.Key,
		Name:    cs.Name,
		Default: cs.Default,
	}
}

// CredentialValue is a simple representation of the actual key/value of a
// Credential.
type CredentialValue struct {
	CredentialSpec `json:",inline"`
	Value          string `json:"value"`
}
