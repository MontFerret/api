package source_test

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/MontFerret/api/source"
)

func TestLocationJSONUsesSemanticSourceName(t *testing.T) {
	value := source.Location{
		Position:   source.Position{Line: 2, Column: 3},
		SourceName: "buffer://query.fql",
	}

	data, err := json.Marshal(value)
	if err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(string(data), `"sourceName":"buffer://query.fql"`) || strings.Contains(string(data), `"file"`) {
		t.Fatalf("unexpected location JSON: %s", data)
	}

	var decoded source.Location
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatal(err)
	}

	if decoded != value {
		t.Fatalf("decoded location = %#v, want %#v", decoded, value)
	}
}
