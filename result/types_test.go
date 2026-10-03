package result_test

import (
	"bytes"
	"encoding/json"
	"reflect"
	"testing"

	"github.com/MontFerret/api/result"
)

func TestContentJSONShapeAndPresence(t *testing.T) {
	for _, tc := range []struct {
		name    string
		content *result.Content
		json    string
	}{
		{name: "absent", json: `null`},
		{name: "zero-valued present", content: &result.Content{}, json: `{"metadata":{"contentType":"","length":0,"lengthKnown":false},"data":null}`},
		{name: "present empty slice", content: &result.Content{Data: []byte{}}, json: `{"metadata":{"contentType":"","length":0,"lengthKnown":false},"data":""}`},
		{name: "known empty with nil data", content: &result.Content{Metadata: result.Metadata{ContentType: "application/json", LengthKnown: true}}, json: `{"metadata":{"contentType":"application/json","length":0,"lengthKnown":true},"data":null}`},
		{name: "known empty with empty slice", content: &result.Content{Metadata: result.Metadata{ContentType: "application/json", LengthKnown: true}, Data: []byte{}}, json: `{"metadata":{"contentType":"application/json","length":0,"lengthKnown":true},"data":""}`},
		{name: "populated", content: &result.Content{Metadata: result.Metadata{ContentType: "application/json", Length: 2, LengthKnown: true}, Data: []byte("42")}, json: `{"metadata":{"contentType":"application/json","length":2,"lengthKnown":true},"data":"NDI="}`},
		{name: "unknown length", content: &result.Content{Metadata: result.Metadata{ContentType: "application/json"}, Data: []byte("42")}, json: `{"metadata":{"contentType":"application/json","length":0,"lengthKnown":false},"data":"NDI="}`},
		{name: "partial content preserves descriptor", content: &result.Content{Metadata: result.Metadata{ContentType: "application/json", Length: 2, LengthKnown: true}, Data: []byte("4")}, json: `{"metadata":{"contentType":"application/json","length":2,"lengthKnown":true},"data":"NA=="}`},
	} {
		t.Run(tc.name, func(t *testing.T) {
			data, err := json.Marshal(tc.content)
			if err != nil {
				t.Fatal(err)
			}
			if string(data) != tc.json {
				t.Fatalf("content JSON = %s, want %s", data, tc.json)
			}
			var decoded *result.Content
			if err := json.Unmarshal(data, &decoded); err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(decoded, tc.content) {
				t.Fatalf("decoded content = %#v, want %#v", decoded, tc.content)
			}
		})
	}
}

func TestContentJSONRoundTripsArbitraryPayloadBytes(t *testing.T) {
	for _, payload := range [][]byte{
		{0, 1, 0x80, 0xff, '{', '"', '\n'},
		[]byte("not a JSON document"),
		[]byte(`{"value":42}`),
	} {
		want := result.Content{
			Metadata: result.Metadata{ContentType: "application/octet-stream", Length: int64(len(payload)), LengthKnown: true},
			Data:     payload,
		}
		data, err := json.Marshal(want)
		if err != nil {
			t.Fatal(err)
		}
		var got result.Content
		if err := json.Unmarshal(data, &got); err != nil {
			t.Fatal(err)
		}
		if got.Metadata != want.Metadata || !bytes.Equal(got.Data, payload) {
			t.Fatalf("round-trip = %#v, want %#v", got, want)
		}
	}
}

func TestMetadataJSONSupportsLengthsBeyond32Bits(t *testing.T) {
	const length int64 = 1 << 40
	want := result.Metadata{ContentType: "application/octet-stream", Length: length, LengthKnown: true}
	data, err := json.Marshal(want)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != `{"contentType":"application/octet-stream","length":1099511627776,"lengthKnown":true}` {
		t.Fatalf("metadata JSON = %s", data)
	}
	var got result.Metadata
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatal(err)
	}
	if got != want {
		t.Fatalf("metadata = %#v, want %#v", got, want)
	}
}
