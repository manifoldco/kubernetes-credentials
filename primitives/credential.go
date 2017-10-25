package primitives

// CredentialSpec represents the specification that is required to filter out
// specific credentials in the Resource spec.
type CredentialSpec struct {
	Key     string `json:"key"`
	Name    string `json:"name,omitempty"`
	Default string `json:"default,omitempty"`
}

// Valid will validate the CredentialSpec.
func (c *CredentialSpec) Valid() bool {
	if c.Key == "" {
		return false
	}

	return true
}

// CredentialValue is a simple representation of the actual key/value of a
// Credential.
type CredentialValue struct {
	CredentialSpec `json:",inline"`
	Value          string `json:"value"`
}
