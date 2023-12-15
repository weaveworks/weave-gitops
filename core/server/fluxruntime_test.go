package server

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

func Test_getRuntimeLabels(t *testing.T) {
	tests := []struct {
		name                     string
		gitopsRuntimeFeatureFlag string
		wantRuntimeLabels        []string
	}{
		{
			name: "should return flux if not feature flag exists",
			wantRuntimeLabels: []string{
				FluxNamespacePartOf,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotLabels := getRuntimeLabels()
			if diff := cmp.Diff(tt.wantRuntimeLabels, gotLabels); diff != "" {
				t.Fatalf("unexpected labels:\n%s", diff)
			}
		})
	}
}
