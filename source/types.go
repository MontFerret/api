package source

type (
	File struct {
		Name    string `json:"name"`
		Content string `json:"content"`
	}

	Position struct {
		Line   int `json:"line"`
		Column int `json:"column"`
	}

	Span struct {
		Start int `json:"start"`
		End   int `json:"end"`
	}

	Location struct {
		Position
		File string `json:"file"`
	}

	Range struct {
		Location
		Span Span `json:"span"`
	}
)
