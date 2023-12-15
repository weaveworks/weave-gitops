package server

import (
	"os"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/weaveworks/weave-gitops/pkg/featureflags"
)

func Test_getRuntimeLabels(t *testing.T) {
	tests := []struct {
		name                     string
		gitopsRuntimeFeatureFlag string
		wantRuntimeLabels        []string
	}{
		{
			name:              "should return flux if not feature flag exists",
			wantRuntimeLabels: FluxRuntimeLabels,
		},
		{
			name:                     "should return flux if feature flag exists but empty",
			wantRuntimeLabels:        FluxRuntimeLabels,
			gitopsRuntimeFeatureFlag: "",
		},
		{
			name:                     "should return flux if feature flag exists and false",
			wantRuntimeLabels:        FluxRuntimeLabels,
			gitopsRuntimeFeatureFlag: "false",
		},
		{
			name:                     "should return weave gitops if feature flag exists and true",
			wantRuntimeLabels:        WeaveGitopsRuntimeLabels,
			gitopsRuntimeFeatureFlag: "true",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.gitopsRuntimeFeatureFlag != "" {
				_ = os.Setenv(GitopsRuntimeFeatureFlag, tt.gitopsRuntimeFeatureFlag)
			}
			defer func() {
				_ = os.Unsetenv(GitopsRuntimeFeatureFlag)
			}()
			featureflags.SetFromEnv(os.Environ())
			gotLabels := getRuntimeLabels()
			if diff := cmp.Diff(tt.wantRuntimeLabels, gotLabels); diff != "" {
				t.Fatalf("unexpected labels:\n%s", diff)
			}
		})
	}
}
