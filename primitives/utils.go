package primitives

import (
	"fmt"

	"k8s.io/api/core/v1"
)

func secretType(t string) (v1.SecretType, error) {
	switch t {
	case "opaque":
	case "":
		return v1.SecretTypeOpaque, nil
	case "docker-registry":
		return v1.SecretTypeDockercfg, nil
	}

	return "", fmt.Errorf("Secret type '%s' not supported", t)
}
