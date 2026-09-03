package source

type (
	// Position identifies a line and column within source text.
	Position struct {
		Line   int `json:"line"`
		Column int `json:"column"`
	}

	// Span identifies a producer-defined interval within source text. Its offset
	// units are defined by the producer rather than by the universal API.
	Span struct {
		Start int `json:"start"`
		End   int `json:"end"`
	}

	// Location adds the semantic source name to a Position. SourceName does not
	// imply that the source is backed by a local filesystem path.
	Location struct {
		Position
		SourceName string `json:"sourceName"`
	}

	// Range combines a source Location with its producer-defined Span.
	Range struct {
		Location
		Span Span `json:"span"`
	}
)
