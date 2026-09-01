package diagnostics_test

import (
	"errors"
	"fmt"
	"reflect"
	"testing"

	"github.com/MontFerret/api/diagnostics"
)

func TestDiagnosticsPreservesOrderAndOptionalAnnotationMessages(t *testing.T) {
	first := diagnostics.Diagnostic{
		Kind:    diagnostics.Kind("SyntaxError"),
		Message: "unexpected token",
	}
	second := diagnostics.Diagnostic{
		Kind:    diagnostics.Unsupported,
		Message: "feature is not supported",
		Annotations: []diagnostics.Annotation{
			{Message: "", Primary: false},
		},
	}

	got := append(diagnostics.Diagnostics{}, first, second)
	want := diagnostics.Diagnostics{first, second}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("Diagnostics = %#v, want %#v", got, want)
	}
	if got[1].Annotations[0].Message != "" || got[1].Annotations[0].Primary {
		t.Fatalf("secondary annotation = %#v, want optional empty message", got[1].Annotations[0])
	}
}

func TestDiagnosticsErrorBehavior(t *testing.T) {
	var nilDiagnostics diagnostics.Diagnostics
	if got := nilDiagnostics.Error(); got != "no diagnostics" {
		t.Fatalf("nil Diagnostics.Error() = %q, want %q", got, "no diagnostics")
	}
	if err := nilDiagnostics.Err(); err != nil {
		t.Fatalf("nil Diagnostics.Err() = %v, want nil", err)
	}

	empty := diagnostics.Diagnostics{}
	if got := empty.Error(); got != "no diagnostics" {
		t.Fatalf("empty Diagnostics.Error() = %q, want %q", got, "no diagnostics")
	}
	if err := empty.Err(); err != nil {
		t.Fatalf("empty Diagnostics.Err() = %v, want nil", err)
	}

	single := diagnostics.Diagnostics{{
		Kind:    diagnostics.UnexpectedError,
		Message: "compilation failed",
	}}
	if got := single.Error(); got != "compilation failed" {
		t.Fatalf("single Diagnostics.Error() = %q, want %q", got, "compilation failed")
	}
	if err := single.Err(); err == nil {
		t.Fatal("single Diagnostics.Err() = nil, want diagnostic failure")
	}

	multiple := diagnostics.Diagnostics{
		{Kind: diagnostics.Kind("SyntaxError"), Message: "unexpected token"},
		{Kind: diagnostics.Kind("NameError"), Message: "unknown binding"},
	}
	if got := multiple.Error(); got != "2 diagnostics" {
		t.Fatalf("multiple Diagnostics.Error() = %q, want %q", got, "2 diagnostics")
	}

	wrapped := fmt.Errorf("compile query: %w", multiple.Err())
	var target diagnostics.Diagnostics
	if !errors.As(wrapped, &target) {
		t.Fatalf("errors.As(%v) did not find Diagnostics", wrapped)
	}
	if !reflect.DeepEqual(target, multiple) {
		t.Fatalf("errors.As target = %#v, want %#v", target, multiple)
	}
}
