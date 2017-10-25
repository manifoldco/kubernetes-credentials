package client

import (
	"fmt"

	"github.com/manifoldco/kubernetes-credentials/primitives"
)

// FlattenResourceCredentialValues will take a set of CredentialValues that are
// linked to a Resource and make a single key/value map. It will use the `Name`
// as key if one is provided and `Default` as value if no value is present and
// a default is set.
func FlattenResourceCredentialValues(resourceCredentials []*primitives.CredentialValue) (map[string]string, error) {
	return FlattenResourcesCredentialValues(
		map[string][]*primitives.CredentialValue{
			"": resourceCredentials,
		},
	)
}

// FlattenResourcesCredentialValues will take a set of CredentialValues that are
// linked to a set Resource and make a single key/value map. It will use the
// `Name` as key if one is provided and `Default` as value if no value is
// present and a default is set.
func FlattenResourcesCredentialValues(resourcesCredentials map[string][]*primitives.CredentialValue) (map[string]string, error) {
	creds := map[string]string{}

	for _, set := range resourcesCredentials {
		for _, cred := range set {
			key := cred.Key
			if cred.Name != "" {
				key = cred.Name
			}

			if _, ok := creds[key]; ok {
				return nil, fmt.Errorf("Key '%s' is already used, please us an alias.", key)
			}

			value := cred.Value
			if value == "" {
				value = cred.Default
			}

			creds[key] = value
		}
	}

	return creds, nil
}
