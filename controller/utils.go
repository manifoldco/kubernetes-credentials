package controller

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"

	"k8s.io/api/core/v1"
)

func secretData(secrets map[string][]byte, secretType v1.SecretType) (map[string][]byte, error) {
	switch secretType {
	case v1.SecretTypeOpaque:
		return secrets, nil
	case v1.SecretTypeDockercfg:
		dockercfg, err := dockerCfg(secrets)
		if err != nil {
			return nil, err
		}

		return map[string][]byte{
			v1.DockerConfigKey: dockercfg,
		}, nil
	default:
		return nil, fmt.Errorf("Secret type '%s' is not supported", string(secretType))
	}
}

const dockerV1Server = "https://index.docker.io/v1/"

// These types represent the docker configuration blocks. See
// https://github.com/kubernetes/kubernetes/blob/84f03ef9572dc6307982d37b78b77621ffa10e37/pkg/credentialprovider/config.go#L33
// for more information.
// This is copied over so we don't have to initialize everything which gets
// initialized in the credentialprovider package.
type (
	dockerConfig      map[string]dockerConfigEntry
	dockerConfigEntry struct {
		Username string
		Password string
		Email    string
	}
	dockerConfigEntryWithAuth struct {
		Username string `json:"username,omitempty"`
		Password string `json:"password,omitempty"`
		Email    string `json:"email,omitempty"`
		Auth     string `json:"auth,omitempty"`
	}
)

func dockerCfg(secrets map[string][]byte) ([]byte, error) {
	server, err := dockerKey(secrets, "server", false)
	if err != nil {
		return nil, err
	}
	if server == "" {
		server = dockerV1Server
	}
	username, err := dockerKey(secrets, "username", true)
	if err != nil {
		return nil, err
	}
	password, err := dockerKey(secrets, "password", true)
	if err != nil {
		return nil, err
	}
	email, err := dockerKey(secrets, "email", true)
	if err != nil {
		return nil, err
	}

	dockercfgAuth := dockerConfigEntry{
		Username: username,
		Password: password,
		Email:    email,
	}

	dockerCfg := dockerConfig{
		server: dockercfgAuth,
	}

	return json.Marshal(dockerCfg)
}

func dockerKey(secrets map[string][]byte, key string, required bool) (string, error) {
	dockerKey := fmt.Sprintf("DOCKER_%s", strings.ToUpper(key))
	value, ok := secrets[dockerKey]
	if required && !ok {
		return "", fmt.Errorf("Expected %s to be set", dockerKey)
	}

	return string(value), nil
}

func (ident *dockerConfigEntry) UnmarshalJSON(data []byte) error {
	var tmp dockerConfigEntryWithAuth
	err := json.Unmarshal(data, &tmp)
	if err != nil {
		return err
	}

	ident.Username = tmp.Username
	ident.Password = tmp.Password
	ident.Email = tmp.Email

	if len(tmp.Auth) == 0 {
		return nil
	}

	ident.Username, ident.Password, err = decodeDockerConfigFieldAuth(tmp.Auth)
	return err
}

func (ident dockerConfigEntry) MarshalJSON() ([]byte, error) {
	toEncode := dockerConfigEntryWithAuth{
		Username: ident.Username,
		Password: ident.Password,
		Email:    ident.Email,
		Auth:     "",
	}
	toEncode.Auth = encodeDockerConfigFieldAuth(ident.Username, ident.Password)

	return json.Marshal(toEncode)
}

// decodeDockerConfigFieldAuth deserializes the "auth" field from dockercfg into a
// username and a password. The format of the auth field is base64(<username>:<password>).
func decodeDockerConfigFieldAuth(field string) (username, password string, err error) {
	decoded, err := base64.StdEncoding.DecodeString(field)
	if err != nil {
		return
	}

	parts := strings.SplitN(string(decoded), ":", 2)
	if len(parts) != 2 {
		err = fmt.Errorf("unable to parse auth field")
		return
	}

	username = parts[0]
	password = parts[1]

	return
}

func encodeDockerConfigFieldAuth(username, password string) string {
	fieldValue := username + ":" + password
	return base64.StdEncoding.EncodeToString([]byte(fieldValue))
}
