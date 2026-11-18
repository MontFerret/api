package debugger

import "errors"

// ErrInvalidValueReference identifies a reference that cannot be resolved in
// the current paused state. Implementations may wrap it. Use errors.Is.
var ErrInvalidValueReference = errors.New("invalid debugger value reference")

// ValueReference identifies an expandable debugger value within one paused
// session state. Resume, termination, and Close invalidate references. An
// expired reference must never identify another value later in that session.
type ValueReference int

// Valid reports whether the reference has a valid portable encoding. It does
// not check whether it belongs to the current paused state.
func (r ValueReference) Valid() bool {
	return r > 0
}
