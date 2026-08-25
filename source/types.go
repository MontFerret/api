package source

type (
	File struct {
		Name    string `json:"name,omitempty"`
		Content string `json:"content,omitempty"`
	}

	Position struct {
		Line   int
		Column int
	}

	Location struct {
		Position
		File string
		Span Span
	}

	Span struct {
		Start int `json:"start"`
		End   int `json:"end"`
	}
)
