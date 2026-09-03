package diagnostics_test

import (
	"encoding/json"
	"reflect"
	"testing"

	"github.com/MontFerret/api/diagnostics"
	"github.com/MontFerret/api/source"
)

func TestDiagnosticPreservesStructuredReportingDetails(t *testing.T) {
	rng := source.Range{
		Location: source.Location{
			Position:   source.Position{Line: 2, Column: 8},
			SourceName: "query.fql",
		},
		Span: source.Span{Start: 19, End: 26},
	}
	diagnostic := diagnostics.Diagnostic{
		Kind:    diagnostics.TypeError,
		Message: "invalid operand type",
		Source:  source.New("query.fql", "let value = 1\nreturn value + true"),
		Annotations: []diagnostics.Annotation{
			{
				Range:   rng,
				Message: "Boolean cannot be added to Int",
				Primary: true,
			},
		},
		Hint: "convert both operands to compatible types",
		Note: "the left operand has type Int",
	}

	if diagnostic.Kind != diagnostics.TypeError {
		t.Fatalf("Kind = %q, want %q", diagnostic.Kind, diagnostics.TypeError)
	}
	if diagnostic.Kind.String() != "TypeError" {
		t.Fatalf("Kind.String() = %q, want %q", diagnostic.Kind.String(), "TypeError")
	}
	if diagnostic.Message != "invalid operand type" {
		t.Fatalf("Message = %q, want %q", diagnostic.Message, "invalid operand type")
	}
	if diagnostic.Source.Name != "query.fql" || diagnostic.Source.Content == "" {
		t.Fatalf("Source = %#v, want named source with content", diagnostic.Source)
	}
	if len(diagnostic.Annotations) != 1 {
		t.Fatalf("len(Annotations) = %d, want 1", len(diagnostic.Annotations))
	}
	if got := diagnostic.Annotations[0]; got.Range != rng || !got.Primary || got.Message != "Boolean cannot be added to Int" {
		t.Fatalf("Annotation = %#v, want primary annotation %#v", got, rng)
	}
	if diagnostic.Hint != "convert both operands to compatible types" {
		t.Fatalf("Hint = %q", diagnostic.Hint)
	}
	if diagnostic.Note != "the left operand has type Int" {
		t.Fatalf("Note = %q", diagnostic.Note)
	}
	if _, ok := any(diagnostic).(error); ok {
		t.Fatal("Diagnostic unexpectedly implements error")
	}
}

func TestDiagnosticJSONRoundTripPreservesSourceRanges(t *testing.T) {
	want := diagnostics.Diagnostic{
		Kind:    diagnostics.Kind("SyntaxError"),
		Message: "expected an expression",
		Source:  source.New("compiler.fql", "let answer =\nreturn answer"),
		Annotations: []diagnostics.Annotation{
			{
				Range: source.Range{
					Location: source.Location{
						Position:   source.Position{Line: 1, Column: 13},
						SourceName: "compiler.fql",
					},
					Span: source.Span{Start: 12, End: 12},
				},
				Message: "expression is missing here",
				Primary: true,
			},
			{
				Range: source.Range{
					Location: source.Location{
						Position:   source.Position{Line: 2, Column: 8},
						SourceName: "compiler.fql",
					},
					Span: source.Span{Start: 20, End: 26},
				},
				Message: "",
				Primary: false,
			},
		},
		Hint: "provide a value after =",
		Note: "the declaration is incomplete",
	}

	data, err := json.Marshal(want)
	if err != nil {
		t.Fatalf("json.Marshal() error = %v", err)
	}

	var got diagnostics.Diagnostic
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("round-trip Diagnostic = %#v, want %#v", got, want)
	}
}
