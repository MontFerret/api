package api_test

import (
	"testing"

	"github.com/MontFerret/api"
)

func TestVersionStringPreservesValue(t *testing.T) {
	for _, tc := range []struct {
		name  string
		value string
	}{
		{name: "empty", value: ""},
		{name: "prerelease", value: "v2.0.0-alpha.55"},
		{name: "release", value: "2.0.0"},
		{name: "development", value: "dev"},
		{name: "unknown", value: "unknown"},
		{name: "opaque with whitespace", value: " \tbuild:abc123+custom\n "},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := api.Version(tc.value).String(); got != tc.value {
				t.Fatalf("Version.String() = %q, want %q", got, tc.value)
			}
		})
	}
}
