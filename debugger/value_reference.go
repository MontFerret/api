package debugger

// ValueReference identifies an expandable debugger value within one paused
// session state. References are invalidated when execution starts or resumes.
type ValueReference int

// Valid reports whether the reference has a valid portable encoding. It does
// not check whether it belongs to the current paused state.
func (r ValueReference) Valid() bool {
	return r > 0
}
