package debugger_test

import (
	"encoding/json"
	"errors"
	"reflect"
	"testing"

	"github.com/MontFerret/api/debugger"
	"github.com/MontFerret/api/result"
)

var _ *result.Content = debugger.Event{}.Output

func TestEventJSONRetainsMaterializedContentAndPresence(t *testing.T) {
	for _, tc := range []struct {
		name    string
		content *result.Content
		json    string
	}{
		{name: "absent", json: `null`},
		{name: "zero-valued present", content: &result.Content{}, json: `{"metadata":{"contentType":"","length":0,"lengthKnown":false},"data":null}`},
		{name: "present empty slice", content: &result.Content{Data: []byte{}}, json: `{"metadata":{"contentType":"","length":0,"lengthKnown":false},"data":""}`},
		{name: "populated", content: &result.Content{Metadata: result.Metadata{ContentType: "application/octet-stream", Length: 2, LengthKnown: true}, Data: []byte{0xff, 0}}, json: `{"metadata":{"contentType":"application/octet-stream","length":2,"lengthKnown":true},"data":"/wA="}`},
	} {
		t.Run(tc.name, func(t *testing.T) {
			want := debugger.Event{Output: tc.content, Reason: debugger.ReasonCompleted}
			data, err := json.Marshal(want)
			if err != nil {
				t.Fatal(err)
			}
			var fields map[string]json.RawMessage
			if err := json.Unmarshal(data, &fields); err != nil {
				t.Fatal(err)
			}
			if string(fields["output"]) != tc.json {
				t.Fatalf("event output JSON = %s, want %s", fields["output"], tc.json)
			}
			if len(fields) != 6 {
				t.Fatalf("event fields = %v, want existing six fields", fields)
			}
			for _, key := range []string{"error", "output", "reason", "hitBreakpointIDs", "location", "depth"} {
				if _, ok := fields[key]; !ok {
					t.Fatalf("event JSON lost key %q", key)
				}
			}
			var got debugger.Event
			if err := json.Unmarshal(data, &got); err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(got, want) {
				t.Fatalf("round-trip event = %#v, want %#v", got, want)
			}
		})
	}
}

func TestEventJSONIncludesContentAlongsideTerminalError(t *testing.T) {
	// The existing error field has no portable JSON error codec. Verify content
	// serialization with a non-nil error without claiming error round-tripping
	// or adapter retention, cloning, or cleanup behavior.
	failure := errors.New("debugger completion cleanup failure")
	content := &result.Content{
		Metadata: result.Metadata{ContentType: "application/json", Length: 2, LengthKnown: true},
		Data:     []byte("4"),
	}
	event := debugger.Event{Output: content, Error: failure, Reason: debugger.ReasonCompleted}
	data, err := json.Marshal(event)
	if err != nil {
		t.Fatal(err)
	}
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(data, &fields); err != nil {
		t.Fatal(err)
	}
	if string(fields["output"]) != `{"metadata":{"contentType":"application/json","length":2,"lengthKnown":true},"data":"NA=="}` || string(fields["error"]) == "null" {
		t.Fatalf("event JSON lost partial content or error: %s", data)
	}
}
