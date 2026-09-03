package debugger

import (
	"github.com/MontFerret/api/result"
	"github.com/MontFerret/api/source"
)

type (
	// PointID identifies a debug point within one compiled program.
	PointID int

	// BreakpointID identifies a breakpoint within one debugger session.
	BreakpointID int

	FunctionID int

	// ValueReference identifies an expandable debugger value within one paused
	// session state. References are invalidated when execution starts or resumes.
	ValueReference int

	// BreakpointBindingMode selects how a requested source location resolves to
	// an executable debug point.
	BreakpointBindingMode int

	// Value is a safely formatted debugger value.
	Value struct {
		Type      string         `json:"type"`
		Display   string         `json:"display"`
		Reference ValueReference `json:"reference"`
	}

	// Variable describes a visible local or bind parameter.
	Variable struct {
		Name    string `json:"name"`
		Value   Value  `json:"value"`
		Mutable bool   `json:"mutable"`
		Param   bool   `json:"param"`
	}

	// Frame describes the paused top frame or one of its callers.
	Frame struct {
		Name       string          `json:"name"`
		Location   source.Location `json:"location"`
		FunctionID FunctionID      `json:"functionID"`
	}

	// Breakpoint describes a requested source-location breakpoint and its resolved
	// executable location, when one exists.
	Breakpoint struct {
		Location          source.Range          `json:"location"`
		RequestedLocation source.Location       `json:"requestedLocation"`
		ID                BreakpointID          `json:"id"`
		PointID           PointID               `json:"pointID"`
		FunctionID        FunctionID            `json:"functionID"`
		BindingMode       BreakpointBindingMode `json:"bindingMode"`
		Bound             bool                  `json:"bound"`
	}

	// BreakpointOptions configures how a requested source location binds.
	BreakpointOptions struct {
		BindingMode BreakpointBindingMode `json:"bindingMode"`
	}

	// Event reports a debugger stop, completion, or termination.
	Event struct {
		Error            error          `json:"error"`
		Output           *result.Output `json:"output"`
		Reason           Reason         `json:"reason"`
		HitBreakpointIDs []BreakpointID `json:"hitBreakpointIDs"`
		Location         source.Range   `json:"location"`
		Depth            int            `json:"depth"`
	}
)

const (
	// BreakpointBindNextExecutableInSource selects the next executable point in
	// the named source and is the zero-value default.
	BreakpointBindNextExecutableInSource BreakpointBindingMode = iota
	BreakpointBindExact
	BreakpointBindNextExecutableInFunction
)

func BreakpointBindingModeFromString(s string) BreakpointBindingMode {
	switch s {
	case "next-executable-in-source":
		return BreakpointBindNextExecutableInSource
	case "exact":
		return BreakpointBindExact
	case "next-executable-in-function":
		return BreakpointBindNextExecutableInFunction
	default:
		return BreakpointBindNextExecutableInSource
	}
}
