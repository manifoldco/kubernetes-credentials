package controller

import (
	"encoding/json"
	"testing"

	"k8s.io/api/core/v1"
)

func TestSecretData(t *testing.T) {
	t.Run("with an opaque type", func(t *testing.T) {
		data := map[string][]byte{
			"key": []byte("value"),
		}

		sData, err := secretData(data, v1.SecretTypeOpaque)
		if err != nil {
			t.Errorf("Expected no error, got '%s'", err)
			t.FailNow()
		}
		if v := string(sData["key"]); v != "value" {
			t.Errorf("Expected value to be 'value', got '%s'", v)
		}
	})

	t.Run("with a docker registry type", func(t *testing.T) {
		t.Run("with a valid set of credentials", func(t *testing.T) {
			data := map[string][]byte{
				"DOCKER_USERNAME": []byte("username"),
				"DOCKER_PASSWORD": []byte("password"),
				"DOCKER_EMAIL":    []byte("email"),
				"DOCKER_SERVER":   []byte("my-server"),
			}

			dData, err := secretData(data, v1.SecretTypeDockercfg)
			if err != nil {
				t.Errorf("Expected no error, got '%s'", err)
				t.FailNow()
			}

			var config dockerConfig
			if err := json.Unmarshal(dData[v1.DockerConfigKey], &config); err != nil {
				t.Errorf("Expected no error unmarshalling the docker config, got '%s'", err)
				t.FailNow()
			}

			auth, ok := config["my-server"]
			if !ok {
				t.Errorf("Expected server auth to be set, not found")
				t.FailNow()
			}

			if auth.Username != "username" {
				t.Errorf("Expected username to be 'username', got '%s'", auth.Username)
			}
			if auth.Password != "password" {
				t.Errorf("Expected password to be 'password', got '%s'", auth.Password)
			}
			if auth.Email != "email" {
				t.Errorf("Expected email to be 'email', got '%s'", auth.Email)
			}
		})

		t.Run("with an invalid set of credentials", func(t *testing.T) {
			data := map[string][]byte{
				"DOCKER_USERNAME": []byte("username"),
				"DOCKER_PASSWORD": []byte("password"),
				"DOCKER_EMAIL":    []byte("email"),
				"DOCKER_SERVER":   []byte("my-server"),
			}

			required := []string{"DOCKER_USERNAME", "DOCKER_PASSWORD", "DOCKER_EMAIL"}
			for _, skip := range required {
				t.Run("missing "+skip, func(t *testing.T) {
					skippedData := map[string][]byte{}
					for k, v := range data {
						if k == skip {
							continue
						}
						skippedData[k] = v
					}

					if _, err := secretData(skippedData, v1.SecretTypeDockercfg); err == nil {
						t.Errorf("Expected error, got none")
					}
				})
			}
		})
	})

	t.Run("with a non-supported type", func(t *testing.T) {
		data := map[string][]byte{
			"key": []byte("value"),
		}

		_, err := secretData(data, v1.SecretTypeServiceAccountToken)
		if err == nil {
			t.Errorf("Expected no error, got none")
		}
	})
}
