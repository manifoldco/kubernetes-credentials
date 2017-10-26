package client_test

import (
	"context"
	"os"
	"testing"

	manifold "github.com/manifoldco/go-manifold"
	"github.com/manifoldco/kubernetes-credentials/helpers/client"
	"github.com/manifoldco/kubernetes-credentials/primitives"
)

var testClient *client.Client

func TestGetResource(t *testing.T) {
	ctx := context.Background()

	t.Run("without a valid resource", func(t *testing.T) {
		invalidResource := &primitives.ResourceSpec{}

		_, err := testClient.GetResource(ctx, nil, invalidResource)
		expectErrorEqual(t, err, client.ErrResourceInvalid)
	})

	t.Run("with a valid resource", func(t *testing.T) {
		t.Run("with a non-existing project", func(t *testing.T) {
			resource := &primitives.ResourceSpec{
				Label: "custom-resource1",
			}

			_, err := testClient.GetResource(ctx, strPtr("non-existing"), resource)
			expectErrorEqual(t, err, client.ErrProjectNotFound)
		})

		t.Run("with an existing project", func(t *testing.T) {
			project := strPtr("kubernetes-secrets")

			t.Run("with an existing resource", func(t *testing.T) {
				resource := &primitives.ResourceSpec{
					Label: "custom-resource1",
				}

				res, err := testClient.GetResource(ctx, project, resource)
				expectNoError(t, err)
				if l := "custom-resource1"; res.Body.Label != l {
					t.Fatalf("Expected label to equal '%s', got '%s'", l, res.Body.Label)
				}
			})

			t.Run("with a non existing resource", func(t *testing.T) {
				resource := &primitives.ResourceSpec{
					Label: "non-existing-resource",
				}

				_, err := testClient.GetResource(ctx, project, resource)
				expectErrorEqual(t, err, client.ErrResourceNotFound)
			})
		})
	})
}

func TestGetResources(t *testing.T) {
	ctx := context.Background()

	t.Run("with an invalid resource", func(t *testing.T) {
		invalidResource := &primitives.ResourceSpec{}

		_, err := testClient.GetResources(ctx, nil, []*primitives.ResourceSpec{invalidResource})
		expectErrorEqual(t, err, client.ErrResourceInvalid)
	})

	t.Run("with valid resources", func(t *testing.T) {
		resources := []*primitives.ResourceSpec{
			{
				Label: "custom-resource1",
			},
			{
				Label: "custom-resource2",
			},
		}

		t.Run("with a non-existing project", func(t *testing.T) {
			_, err := testClient.GetResources(ctx, strPtr("non-existing"), resources)
			expectErrorEqual(t, err, client.ErrProjectNotFound)
		})

		t.Run("with an existing project", func(t *testing.T) {
			project := strPtr("kubernetes-secrets")

			t.Run("with one non-existing resource", func(t *testing.T) {
				nonExisting := &primitives.ResourceSpec{
					Label: "non-existing",
				}
				nr := append(resources, nonExisting)
				_, err := testClient.GetResources(ctx, project, nr)
				expectErrorEqual(t, err, client.ErrResourceNotFound)
			})

			t.Run("with all existing resources", func(t *testing.T) {
				res, err := testClient.GetResources(ctx, project, resources)
				expectNoError(t, err)
				if len(res) != 2 {
					t.Fatalf("Expected '2' resources to be loaded, got '%d'", len(res))
				}
			})
		})
	})
}

func TestGetResourceCredentialValues(t *testing.T) {
	ctx := context.Background()

	t.Run("with an invalid resource", func(t *testing.T) {
		invalidResource := &primitives.ResourceSpec{}

		_, err := testClient.GetResourceCredentialValues(ctx, nil, invalidResource)
		expectErrorEqual(t, err, client.ErrResourceInvalid)
	})

	t.Run("with a valid resource", func(t *testing.T) {
		res := &primitives.ResourceSpec{
			Label: "custom-resource1",
		}

		t.Run("with a non-existing project", func(t *testing.T) {
			_, err := testClient.GetResourceCredentialValues(ctx, strPtr("non-existing"), res)
			expectErrorEqual(t, err, client.ErrProjectNotFound)
		})

		t.Run("with an existing project", func(t *testing.T) {
			project := strPtr("kubernetes-secrets")

			t.Run("without credentials subset", func(t *testing.T) {
				creds, err := testClient.GetResourceCredentialValues(ctx, project, res)
				expectNoError(t, err)

				if len(creds) != 2 {
					t.Fatalf("Expected '2' CredentialValues for 'custom-resource1', got '%d'", len(creds))
				}
				data, err := client.FlattenResourceCredentialValues(creds)
				expectNoError(t, err)

				expectStringEqual(t, data["TOKEN_ID"], "my-secret-token-id")
				expectStringEqual(t, data["TOKEN_SECRET"], "my-secret-token-secret")
			})

			t.Run("with a valid credential subset", func(t *testing.T) {
				sub := &primitives.ResourceSpec{
					Label: "custom-resource1",
					Credentials: []*primitives.CredentialSpec{
						{
							Key: "TOKEN_ID",
						},
					},
				}

				creds, err := testClient.GetResourceCredentialValues(ctx, project, sub)
				expectNoError(t, err)
				if len(creds) != 1 {
					t.Fatalf("Expected '1' CredentialValues for 'custom-resource1', got '%d'", len(creds))
				}

				data, err := client.FlattenResourceCredentialValues(creds)
				expectNoError(t, err)
				expectStringEqual(t, data["TOKEN_ID"], "my-secret-token-id")
			})

			t.Run("with a non existing key", func(t *testing.T) {
				t.Run("with a default value", func(t *testing.T) {
					sub := &primitives.ResourceSpec{
						Label: "custom-resource1",
						Credentials: []*primitives.CredentialSpec{
							{
								Key:     "NON_EXISTING",
								Default: "my-default-value",
							},
						},
					}

					creds, err := testClient.GetResourceCredentialValues(ctx, project, sub)
					expectNoError(t, err)
					if len(creds) != 1 {
						t.Fatalf("Expected '1' CredentialValues for 'custom-resource1', got '%d'", len(creds))
					}

					data, err := client.FlattenResourceCredentialValues(creds)
					expectNoError(t, err)
					expectStringEqual(t, data["NON_EXISTING"], "my-default-value")
				})

				t.Run("without a default value", func(t *testing.T) {
					sub := &primitives.ResourceSpec{
						Label: "custom-resource1",
						Credentials: []*primitives.CredentialSpec{
							{
								Key: "NON_EXISTING",
							},
						},
					}

					_, err := testClient.GetResourceCredentialValues(ctx, project, sub)
					expectErrorEqual(t, err, client.ErrCredentialDefaultNotSet)
				})
			})

			t.Run("with an invalid credential subset", func(t *testing.T) {
				sub := &primitives.ResourceSpec{
					Label: "custom-resource1",
					Credentials: []*primitives.CredentialSpec{
						{
							Name: "Invalid",
						},
					},
				}

				_, err := testClient.GetResourceCredentialValues(ctx, project, sub)
				expectErrorEqual(t, err, client.ErrResourceInvalid)
			})
		})
	})
}

func init() {
	testClient = newClient()
}

func newClient() *client.Client {
	c, err := client.New(
		manifold.New(
			manifold.WithAPIToken(os.Getenv("MANIFOLD_API_TOKEN")),
		),
		strPtr(os.Getenv("MANIFOLD_TEAM")),
	)

	if err != nil {
		panic("Could not set up the test client: " + err.Error())
	}

	return c
}

func expectErrorEqual(t *testing.T, act, exp error) {
	if act == nil {
		t.Fatalf("Expected error not to be 'nil' but '%s'", exp.Error())
	}

	if exp != act {
		t.Fatalf("Expected error '%s', to equal '%s'", act.Error(), exp.Error())
	}
}

func expectNoError(t *testing.T, err error) {
	if err != nil {
		t.Fatalf("Expected no error to have occurred, got '%s'", err)
	}
}

func expectStringEqual(t *testing.T, act, exp string) {
	if act != exp {
		t.Fatalf("Expected '%s' to equal '%s'", act, exp)
	}
}

func strPtr(str string) *string {
	return &str
}
