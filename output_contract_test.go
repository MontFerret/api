package api_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/MontFerret/api"
	"github.com/MontFerret/api/debugger"
	"github.com/MontFerret/api/result"
)

// Method expressions catch pointer-to-interface execution signatures and stale
// buffered return types. Assignments in both directions require exact aliases.
var (
	_ func(api.Runtime, context.Context, api.Source, ...api.SessionOption) (api.Output, error) = api.Runtime.Run
	_ func(api.Session, context.Context) (api.Output, error)                                   = api.Session.Run
	_ func(api.Output) api.Metadata                                                            = api.Output.Metadata
	_ func(api.Output, context.Context, api.Consumer) error                                    = api.Output.Consume
	_ func(api.Output, context.Context) (*api.Content, error)                                  = api.Output.Collect
	_ func(api.Output) error                                                                   = api.Output.Close
	_ func(result.Output, context.Context, result.Consumer) error                              = api.Output.Consume
	_ func(api.Output, context.Context) (*api.Content, error)                                  = result.Output.Collect
	_ api.Metadata                                                                             = result.Metadata{}
	_ result.Metadata                                                                          = api.Metadata{}
	_ *api.Content                                                                             = (*result.Content)(nil)
	_ *result.Content                                                                          = (*api.Content)(nil)
	_ api.Consumer                                                                             = result.Consumer(nil)
	_ result.Consumer                                                                          = api.Consumer(nil)
	_ func(context.Context, []byte) error                                                      = api.Consumer(nil)
	_ *api.Content                                                                             = debugger.Event{}.Output
)

// The module declares interfaces only. This matrix verifies the collection
// example's caller behavior with scripted outcomes, not adapter conformance.
func TestCollectQueryPreservesContentPresenceAndErrors(t *testing.T) {
	failure := errors.New("terminal execution failure")
	cleanupFailure := errors.New("output cleanup failure")
	populated := &api.Content{
		Metadata: api.Metadata{ContentType: "application/json", Length: 2, LengthKnown: true},
		Data:     []byte("42"),
	}
	prefix := &api.Content{Metadata: populated.Metadata, Data: []byte("4")}

	for _, tc := range []struct {
		name       string
		content    *api.Content
		terminal   error
		cleanupErr error
	}{
		{name: "absent"},
		{name: "absent with error", terminal: failure},
		{name: "zero-valued present", content: &api.Content{}},
		{name: "zero-valued present with error", content: &api.Content{}, terminal: failure},
		{name: "present empty slice", content: &api.Content{Data: []byte{}}},
		{name: "populated", content: populated},
		{name: "populated with error", content: populated, terminal: failure},
		{name: "prefix with error", content: prefix, terminal: failure},
		{name: "cleanup failure", content: populated, cleanupErr: cleanupFailure},
		{name: "multiple failures", content: prefix, terminal: failure, cleanupErr: cleanupFailure},
	} {
		t.Run(tc.name, func(t *testing.T) {
			output := &scriptedOutput{content: tc.content, terminalErr: tc.terminal, cleanupErr: tc.cleanupErr}
			content, err := collectQuery(t.Context(), &outputRuntime{output: output}, api.NewAnonymousSource("RETURN 42"))
			if content != tc.content {
				t.Fatalf("content = %p, want available content %p", content, tc.content)
			}
			if tc.terminal == nil && tc.cleanupErr == nil && err != nil {
				t.Fatalf("unexpected collection error: %v", err)
			}
			for _, cause := range []error{tc.terminal, tc.cleanupErr} {
				if cause != nil && !errors.Is(err, cause) {
					t.Fatalf("collection error %v lost cause %v", err, cause)
				}
			}
			if output.collectCalls != 1 || output.consumeCalls != 0 || output.closeCalls != 1 {
				t.Fatalf("caller calls: collect=%d consume=%d fallback close=%d", output.collectCalls, output.consumeCalls, output.closeCalls)
			}
		})
	}
}

func TestOutputSentinelsShareIdentityAndMatchThroughTraversal(t *testing.T) {
	if api.ErrOutputInUse != result.ErrInUse || api.ErrOutputClosed != result.ErrClosed {
		t.Fatal("root exports must reference the canonical result sentinels")
	}
	other := errors.New("another observed failure")
	for _, sentinel := range []error{api.ErrOutputInUse, api.ErrOutputClosed} {
		for _, err := range []error{sentinel, fmt.Errorf("consume: %w", sentinel), errors.Join(other, fmt.Errorf("collect: %w", sentinel))} {
			if !errors.Is(err, sentinel) {
				t.Fatalf("errors.Is(%v, %v) = false", err, sentinel)
			}
		}
		if errors.Is(errors.New(sentinel.Error()), sentinel) {
			t.Fatal("matching text must not substitute for sentinel identity")
		}
	}
	joined := errors.Join(api.ErrOutputClosed, other)
	if !errors.Is(joined, api.ErrOutputClosed) || !errors.Is(joined, other) {
		t.Fatal("joined closure failure lost an observed cause")
	}
	if errors.Is(api.ErrOutputInUse, api.ErrOutputClosed) || errors.Is(api.ErrOutputClosed, api.ErrOutputInUse) {
		t.Fatal("in-use and closed sentinels must be distinct")
	}
}
