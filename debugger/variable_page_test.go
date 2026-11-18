package debugger_test

import (
	"encoding/json"
	"errors"
	"math"
	"reflect"
	"testing"

	"github.com/MontFerret/api/debugger"
)

func TestVariablePageNormalize(t *testing.T) {
	for _, tc := range []struct {
		name string
		page debugger.VariablePage
		want int64
	}{
		{"default", debugger.VariablePage{}, debugger.DefaultVariablePageSize},
		{"indexed", debugger.VariablePage{Filter: debugger.VariableFilterIndexed, Start: 1000, Count: 20}, 20},
		{"maximum", debugger.VariablePage{Filter: debugger.VariableFilterNamed, Count: debugger.MaxVariablePageSize}, debugger.MaxVariablePageSize},
		{"last representable end", debugger.VariablePage{Start: math.MaxInt64 - 1, Count: 1}, 1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got, err := tc.page.Normalize()
			want := tc.page
			want.Count = tc.want
			if err != nil || got != want {
				t.Fatalf("Normalize() = %+v, %v, want %+v", got, err, want)
			}
		})
	}
}

func TestVariablePageRejectsInvalidRangesAndFilters(t *testing.T) {
	for _, page := range []debugger.VariablePage{
		{Filter: "other"},
		{Start: -1},
		{Count: -1},
		{Count: math.MinInt64},
		{Count: debugger.MaxVariablePageSize + 1},
		{Start: math.MaxInt64, Count: 1},
		{Start: math.MaxInt64 - debugger.DefaultVariablePageSize + 1},
	} {
		if _, err := page.Normalize(); !errors.Is(err, debugger.ErrInvalidVariablePage) {
			t.Errorf("Normalize(%+v) error = %v, want ErrInvalidVariablePage", page, err)
		}
	}
}

func TestValueJSONPreservesChildCounts(t *testing.T) {
	var zero int64
	large := int64(1 << 40)
	for _, tc := range []struct {
		name    string
		value   debugger.Value
		indexed string
		named   string
	}{
		{"unknown counts", debugger.Value{Type: "Collection", Reference: 1}, "", ""},
		{"empty array", debugger.Value{Type: "Array", Display: "Array(0)", IndexedCount: &zero}, "0", ""},
		{"indexed children", debugger.Value{Reference: 1, IndexedCount: &large, NamedCount: &zero}, "1099511627776", "0"},
		{"summary-only object", debugger.Value{Type: "Object", NamedCount: &large}, "", "1099511627776"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			data, err := json.Marshal(tc.value)
			if err != nil {
				t.Fatal(err)
			}
			var fields map[string]json.RawMessage
			if err := json.Unmarshal(data, &fields); err != nil {
				t.Fatal(err)
			}
			if string(fields["indexedCount"]) != tc.indexed || string(fields["namedCount"]) != tc.named {
				t.Fatalf("unexpected child count presence or value in %s", data)
			}
			var got debugger.Value
			if err := json.Unmarshal(data, &got); err != nil || !reflect.DeepEqual(got, tc.value) {
				t.Fatalf("round trip of %s = %+v, %v", data, got, err)
			}
		})
	}
}

func TestVariablePageJSON(t *testing.T) {
	fields := func(data []byte) map[string]json.RawMessage {
		t.Helper()
		var obj map[string]json.RawMessage
		if err := json.Unmarshal(data, &obj); err != nil {
			t.Fatal(err)
		}
		return obj
	}
	for _, tc := range []struct {
		wire string
		page debugger.VariablePage
	}{
		{`{}`, debugger.VariablePage{}},
		{`{"filter":"indexed","start":1099511627776,"count":20}`, debugger.VariablePage{Filter: debugger.VariableFilterIndexed, Start: 1 << 40, Count: 20}},
		{`{"filter":"named"}`, debugger.VariablePage{Filter: debugger.VariableFilterNamed}},
	} {
		var got debugger.VariablePage
		if err := json.Unmarshal([]byte(tc.wire), &got); err != nil || got != tc.page {
			t.Errorf("decode %s = %+v, %v, want %+v", tc.wire, got, err, tc.page)
		}
		data, err := json.Marshal(tc.page)
		if err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(fields(data), fields([]byte(tc.wire))) {
			t.Errorf("encode %+v = %s, want fields of %s", tc.page, data, tc.wire)
		}
	}
}

func TestVariableJSONPreservesExactChildNames(t *testing.T) {
	for _, name := range []string{"", "before\x00after", "\uFFFD", "\u00E9", "e\u0301"} {
		data, err := json.Marshal(debugger.Variable{Name: name})
		if err != nil {
			t.Fatal(err)
		}
		// An omitted name must be distinguishable from a present empty key.
		got := debugger.Variable{Name: "missing name"}
		if err := json.Unmarshal(data, &got); err != nil {
			t.Fatal(err)
		}
		if got.Name != name {
			t.Errorf("child name after JSON round trip = %q, want %q", got.Name, name)
		}
	}
}
