package primitives

import "testing"

func TestProjectSpec_String(t *testing.T) {

	tcs := []struct {
		scenario string
		spec     ProjectSpec
		result   string
	}{
		{
			scenario: "empty",
			result:   "",
		},
		{
			scenario: "with name",
			spec: ProjectSpec{
				Name: "production",
			},
			result: "project: production",
		},
		{
			scenario: "with team",
			spec: ProjectSpec{
				Team: "manifold",
			},
			result: "team: manifold",
		},
		{
			scenario: "with type",
			spec: ProjectSpec{
				Type: "docker-registry",
			},
			result: "type: docker-registry",
		},
		{
			scenario: "with all fields",
			spec: ProjectSpec{
				Name: "production",
				Team: "manifold",
				Type: "docker-registry",
			},
			result: "project: production, team: manifold, type: docker-registry",
		},
	}

	for _, tc := range tcs {
		t.Run(tc.scenario, func(t *testing.T) {
			got := tc.spec.String()

			if got != tc.result {
				t.Fatalf("expected project spec string to eq %q, got %q", tc.result, got)
			}
		})
	}
}
