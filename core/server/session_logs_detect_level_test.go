package server

import (
	"testing"
)

func TestDetectLogLevel(t *testing.T) {
	testCases := []struct {
		message  string
		expected string
	}{
		{"[ERROR] something went wrong", "error"},
		{"[warn] disk is almost full", "warn"},
		{"[info] system started", "info"},
		{"[ftl] system failure", "error"},
		{"application error", "error"},
		{"application warning", "warn"},
		{"application info", "info"},
	}
	for _, tc := range testCases {
		actual := detectLogLevel(tc.message)
		if actual != tc.expected {
			t.Errorf("For message %s, expected log level to be %s, but got %s", tc.message, tc.expected, actual)
		}
	}
}
