package run

import "testing"

func TestTrimK8sVersion(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string
	}{
		{
			name: "1.18.0",
			in:   "1.18.0",
			want: "1.18.0",
		},
		{
			name: "v1.18.0",
			in:   "v1.18.0",
			want: "1.18.0",
		},
		{
			name: "v1.18.0+",
			in:   "v1.18.0+",
			want: "1.18.0",
		},
		{
			name: "v1.18.0+123",
			in:   "v1.18.0+123",
			want: "1.18.0",
		},
		{
			name: "v1.18.0+123-abc",
			in:   "v1.18.0+123-abc",
			want: "1.18.0",
		},
		{
			name: "v1.18.0-abc",
			in:   "v1.18.0-abc",
			want: "1.18.0",
		},
		{
			name: "v1.18.0-abc+123",
			in:   "v1.18.0-abc+123",
			want: "1.18.0",
		},
		{
			name: "v1.18.0-abc+123-def",
			in:   "v1.18.0-abc+123-def",
			want: "1.18.0",
		},
		{
			name: "v1.18.0-abc+123-def+456",
			in:   "v1.18.0-abc+123-def+456",
			want: "1.18.0",
		},
		{
			name: "v1.18.0-abc+123-def+456-ghi",
			in:   "v1.18.0-abc+123-def+456-ghi",
			want: "1.18.0",
		},
		{
			name: "v1.18.0-abc+123-def+456-ghi+jkl",
			in:   "v1.18.0-abc+123-def+456-ghi+jkl",
			want: "1.18.0",
		},
		{
			name: "1.18.0-abc+123-def+456-ghi+jkl",
			in:   "1.18.0-abc+123-def+456-ghi+jkl",
			want: "1.18.0",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := trimK8sVersion(tt.in); got != tt.want {
				t.Errorf("trimK8sVersion() = %v, want %v", got, tt.want)
			}
		})
	}
}
