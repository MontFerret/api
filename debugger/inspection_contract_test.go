package debugger_test

import (
	"context"
	"errors"
	"fmt"
	"math"
	"slices"
	"strconv"
	"testing"

	"github.com/MontFerret/api/debugger"
)

// inspectionFixture supplies a paused binding and scripted page reads to caller
// examples. The runtime owns handle allocation and stop transitions.
type inspectionFixture struct {
	sessionContract
	value debugger.Value
	read  func(debugger.VariablePage) ([]debugger.Variable, error)
}

func (s inspectionFixture) Locals(context.Context) ([]debugger.Variable, error) {
	return []debugger.Variable{{Name: "results", Value: s.value}}, nil
}

func (s inspectionFixture) FrameLocals(ctx context.Context, frame int) ([]debugger.Variable, error) {
	if frame != 0 {
		return nil, fmt.Errorf("fixture has no frame %d", frame)
	}
	return s.Locals(ctx)
}

func (s inspectionFixture) Variables(ctx context.Context, ref debugger.ValueReference, page debugger.VariablePage) ([]debugger.Variable, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if !ref.Valid() || ref != s.value.Reference {
		return nil, debugger.ErrInvalidValueReference
	}
	page, err := page.Normalize()
	if err != nil {
		return nil, err
	}
	return s.read(page)
}

func TestInspectionReadsOnlyTheRequestedMillionElementArrayPage(t *testing.T) {
	length, noNames := int64(1_000_000), int64(0)
	var reads []int64
	var pageCalls int
	var session debugger.Session = inspectionFixture{
		value: debugger.Value{Type: "Array", Display: "Array(1000000)", Reference: 1, IndexedCount: &length, NamedCount: &noNames},
		read: func(page debugger.VariablePage) ([]debugger.Variable, error) {
			pageCalls++
			if page.Filter == debugger.VariableFilterNamed {
				return nil, nil
			}
			var children []debugger.Variable
			for index := page.Start; index < min(length, page.Start+page.Count); index++ {
				reads = append(reads, index)
				children = append(children, debugger.Variable{Name: strconv.FormatInt(index, 10)})
			}
			return children, nil
		},
	}
	locals, err := session.Locals(t.Context())
	if err != nil || len(locals) != 1 || len(reads) != 0 {
		t.Fatalf("Locals = %v, %v; child reads = %v", locals, err, reads)
	}
	value := locals[0].Value
	if !value.Reference.Valid() {
		t.Fatalf("array reference = %d", value.Reference)
	}
	topFrame := func(ctx context.Context) ([]debugger.Variable, error) { return session.FrameLocals(ctx, 0) }
	for _, list := range []func(context.Context) ([]debugger.Variable, error){topFrame, session.Locals, topFrame} {
		bindings, err := list(t.Context())
		if err != nil || len(bindings) != 1 || pageCalls != 0 {
			t.Fatalf("relisted bindings = %v, %v; page calls = %d", bindings, err, pageCalls)
		}
		if binding := bindings[0]; binding.Name != locals[0].Name || binding.Value.Reference != value.Reference {
			t.Fatalf("relisted binding = %v, want %s with reference %d", binding, locals[0].Name, value.Reference)
		}
	}
	if value.IndexedCount == nil || *value.IndexedCount != 1_000_000 || value.NamedCount == nil || *value.NamedCount != 0 {
		t.Fatalf("array counts = %+v", value)
	}
	named, err := session.Variables(t.Context(), value.Reference, debugger.VariablePage{Filter: debugger.VariableFilterNamed})
	if err != nil || len(named) != 0 || len(reads) != 0 {
		t.Fatalf("absent named children = %v, %v; child reads = %v", named, err, reads)
	}
	children, err := session.Variables(t.Context(), value.Reference, debugger.VariablePage{Filter: debugger.VariableFilterIndexed, Start: 1000, Count: 20})
	if err != nil || len(children) != 20 || len(reads) != 20 {
		t.Fatalf("page length = %d, reads = %v, error = %v", len(children), reads, err)
	}
	for i, child := range children {
		if child.Name != strconv.Itoa(1000+i) || reads[i] != int64(1000+i) {
			t.Fatalf("child %d = %q, source index = %d", i, child.Name, reads[i])
		}
	}
}

func TestInspectionPagesWithUnknownCounts(t *testing.T) {
	// The caller only knows the reference. Each slice describes the source order.
	views := map[debugger.VariableFilter][]debugger.Variable{
		debugger.VariableFilterNamed:   {{Name: "0"}, {Name: "size"}},
		debugger.VariableFilterIndexed: {{Name: "0"}, {Name: "1"}, {Name: "2"}, {Name: "3"}},
	}
	views[debugger.VariableFilterAll] = slices.Concat(views[debugger.VariableFilterNamed], views[debugger.VariableFilterIndexed])
	session := inspectionFixture{value: debugger.Value{Reference: 1}, read: func(page debugger.VariablePage) ([]debugger.Variable, error) {
		children := views[page.Filter]
		end := int64(len(children))
		return children[min(page.Start, end):min(page.Start+page.Count, end)], nil
	}}
	for _, tc := range []struct {
		page debugger.VariablePage
		want []string
	}{
		{debugger.VariablePage{Count: 3}, []string{"0", "size", "0"}},
		{debugger.VariablePage{Start: 1, Count: 3}, []string{"size", "0", "1"}},
		{debugger.VariablePage{Start: 4, Count: 2}, []string{"2", "3"}},
		{debugger.VariablePage{Filter: debugger.VariableFilterNamed, Start: 1, Count: 4}, []string{"size"}},
		{debugger.VariablePage{Filter: debugger.VariableFilterIndexed, Start: 3, Count: 2}, []string{"3"}},
		{debugger.VariablePage{Start: 6}, nil},
		{debugger.VariablePage{Start: 8}, nil},
	} {
		got, err := session.Variables(t.Context(), session.value.Reference, tc.page)
		if err != nil || len(got) != len(tc.want) {
			t.Fatalf("Variables(%+v) = %v, %v; want %v", tc.page, got, err, tc.want)
		}
		for i, child := range got {
			if child.Name != tc.want[i] {
				t.Errorf("page %+v child %d = %q, want %q", tc.page, i, child.Name, tc.want[i])
			}
		}
	}
}

func TestInspectionContinuesAtSignedBoundaryWithIndexedFilter(t *testing.T) {
	// One named child precedes MaxInt64 indexed children in the combined view.
	const last = math.MaxInt64 - 1
	var reads []int64
	session := inspectionFixture{value: debugger.Value{Reference: 1}, read: func(page debugger.VariablePage) ([]debugger.Variable, error) {
		index := page.Start
		if page.Filter == debugger.VariableFilterAll {
			index--
		}
		reads = append(reads, index)
		return []debugger.Variable{{Name: strconv.FormatInt(index, 10)}}, nil
	}}
	page := debugger.VariablePage{Start: last, Count: 1}
	got, err := session.Variables(t.Context(), 1, page)
	if err != nil || len(got) != 1 || got[0].Name != strconv.FormatInt(last-1, 10) {
		t.Fatalf("last combined page = %v, %v", got, err)
	}
	got, err = session.Variables(t.Context(), 1, debugger.VariablePage{Start: math.MaxInt64, Count: 1})
	if got != nil || !errors.Is(err, debugger.ErrInvalidVariablePage) || !slices.Equal(reads, []int64{last - 1}) {
		t.Fatalf("combined continuation = %v, %v; reads = %v", got, err, reads)
	}
	page.Filter = debugger.VariableFilterIndexed
	got, err = session.Variables(t.Context(), 1, page)
	if err != nil || len(got) != 1 || got[0].Name != strconv.FormatInt(last, 10) || !slices.Equal(reads, []int64{last - 1, last}) {
		t.Fatalf("indexed continuation = %v, %v; reads = %v", got, err, reads)
	}
}

func TestInspectionUsesIndexedViewAfterUnsupportedCombinedRange(t *testing.T) {
	combined := debugger.VariablePage{Start: 1000, Count: 2}
	indexed := debugger.VariablePage{Filter: debugger.VariableFilterIndexed, Start: 0, Count: 2}
	session := inspectionFixture{value: debugger.Value{Reference: 1}, read: func(page debugger.VariablePage) ([]debugger.Variable, error) {
		switch page {
		case combined:
			return nil, fmt.Errorf("cannot locate the named/indexed boundary: %w", errors.ErrUnsupported)
		case indexed:
			return []debugger.Variable{{Name: "0"}, {Name: "1"}}, nil
		default:
			return nil, fmt.Errorf("unexpected page %+v", page)
		}
	}}
	got, err := session.Variables(t.Context(), session.value.Reference, combined)
	if got != nil || !errors.Is(err, errors.ErrUnsupported) {
		t.Fatalf("combined page = %v, %v; want nil and unsupported", got, err)
	}
	got, err = session.Variables(t.Context(), session.value.Reference, indexed)
	if err != nil || len(got) != 2 || got[0].Name != "0" || got[1].Name != "1" {
		t.Fatalf("first indexed page = %v, %v", got, err)
	}
}

func TestInspectionErrorResults(t *testing.T) {
	// This stop issued reference 2. Reference 1 belonged to the previous stop.
	session := inspectionFixture{value: debugger.Value{Reference: 2}}
	for _, ref := range []debugger.ValueReference{-1, 0, 1, 99} {
		got, err := session.Variables(t.Context(), ref, debugger.VariablePage{})
		if got != nil || !errors.Is(err, debugger.ErrInvalidValueReference) {
			t.Fatalf("Variables(%d) = %v, %v", ref, got, err)
		}
	}
	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	page := debugger.VariablePage{Count: 1}
	got, err := session.Variables(ctx, 2, page)
	if got != nil || !errors.Is(err, context.Canceled) {
		t.Fatalf("canceled page = %v, %v", got, err)
	}
	readErr := errors.New("child read failed")
	session.read = func(debugger.VariablePage) ([]debugger.Variable, error) { return nil, readErr }
	got, err = session.Variables(t.Context(), 2, page)
	if got != nil || !errors.Is(err, readErr) {
		t.Fatalf("failed page = %v, %v", got, err)
	}
	session.read = func(debugger.VariablePage) ([]debugger.Variable, error) {
		return []debugger.Variable{{Name: "0"}}, nil
	}
	got, err = session.Variables(t.Context(), 2, page)
	if err != nil || len(got) != 1 || got[0].Name != "0" {
		t.Fatalf("retry at same stop = %v, %v", got, err)
	}
}
