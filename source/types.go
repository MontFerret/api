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

	// Location adds the source file identity to a Position.
	Location struct {
		Position
		File string `json:"file"`
	}

	// Range combines a source Location with its producer-defined Span.
	Range struct {
		Location
		Span Span `json:"span"`
	}
)
