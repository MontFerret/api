package diagnostics

import "github.com/MontFerret/api/source"

type (
	// Kind classifies a diagnostic. It is open so diagnostic producers can use
	// domain-specific kinds without changing the universal API.
	Kind string

	// Annotation describes one highlighted source range associated with a
	// diagnostic. Message may be empty when the range needs no separate label.
	Annotation struct {
		Range   source.Range `json:"range"`
		Message string       `json:"message"`
		Primary bool         `json:"primary"`
	}

	// Diagnostic describes one portable structured issue reported by a runtime
	// or tool. It intentionally carries no implementation-specific error cause.
	Diagnostic struct {
		Kind        Kind          `json:"kind"`
		Message     string        `json:"message"`
		Source      source.Source `json:"source"`
		Annotations []Annotation  `json:"annotations"`
		Hint        string        `json:"hint"`
		Note        string        `json:"note"`
	}
)

// Shared diagnostic kinds. Producers may define additional Kind values in
// their owning packages.
const (
	Unknown         Kind = ""
	Unsupported     Kind = "Unsupported"
	UnexpectedError Kind = "UnexpectedError"
	TypeError       Kind = "TypeError"
)

// String returns the diagnostic kind as text.
func (k Kind) String() string {
	return string(k)
}
