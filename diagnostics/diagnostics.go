package diagnostics

import "fmt"

// Diagnostics is an ordered portable collection of Diagnostic values. It is
// the error-bearing diagnostic type, including for failures with one item.
type Diagnostics []Diagnostic

// Error returns a concise summary derived from the collection.
func (d Diagnostics) Error() string {
	switch len(d) {
	case 0:
		return "no diagnostics"
	case 1:
		return d[0].Message
	default:
		return fmt.Sprintf("%d diagnostics", len(d))
	}
}

// Err returns nil for an empty collection and the collection otherwise. The
// returned concrete error is Diagnostics, so wrapped diagnostic failures can
// be detected with errors.As into a Diagnostics value.
func (d Diagnostics) Err() error {
	if len(d) == 0 {
		return nil
	}

	return d
}
