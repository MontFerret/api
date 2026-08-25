package debugger

import (
	"context"
	"io"

	"github.com/MontFerret/api/result"
	"github.com/MontFerret/api/source"
)

type (
	// Reason identifies why a debug execution stopped.
	Reason string

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
		RequestedPosition source.Location       `json:"requestedPosition"`
		ID                BreakpointID          `json:"id"`
		PointID           PointID               `json:"pointID"`
		FunctionID        FunctionID            `json:"functionID"`
		BindingMode       BreakpointBindingMode `json:"bindingMode"`
		Bound             bool
	}

	// BreakpointOptions configures how a requested source location binds.
	BreakpointOptions struct {
		BindingMode BreakpointBindingMode `json:"bindingMode"`
	}

	// Event reports a debugger stop, completion, or termination.
	Event struct {
		Error            error          `json:"error"`
		Output           result.Output  `json:"output"`
		Reason           Reason         `json:"reason"`
		HitBreakpointIDs []BreakpointID `json:"hitBreakpointIDs"`
		Location         source.Range   `json:"location"`
		Depth            int            `json:"depth"`
	}

	Session interface {
		io.Closer
		Start(ctx context.Context) (*Event, error)
		Continue(ctx context.Context) (*Event, error)
		Step(ctx context.Context) (*Event, error)
		Next(ctx context.Context) (*Event, error)
		Out(ctx context.Context) (*Event, error)
		Pause() error
		SetBreakpoint(pos source.Location) (Breakpoint, error)
		SetBreakpointAt(loc source.Location, opts BreakpointOptions) (Breakpoint, error)
		DeleteBreakpoint(id BreakpointID) error
		Breakpoints() []Breakpoint
		Frames() ([]Frame, error)
		Locals() ([]Variable, error)
		FrameLocals(frame int) ([]Variable, error)
		Variables(reference ValueReference) ([]Variable, error)
		Evaluate(ctx context.Context, expression string) (Value, error)
		EvaluateFrame(ctx context.Context, frame int, expression string) (Value, error)
	}
)

const (
	ReasonEntry        Reason = "entry"
	ReasonBreakpoint   Reason = "breakpoint"
	ReasonStep         Reason = "step"
	ReasonPause        Reason = "pause"
	ReasonRuntimeError Reason = "runtime-error"
	ReasonCompleted    Reason = "completed"
	ReasonTerminated   Reason = "terminated"
)

const (
	// BreakpointBindNextExecutableInFile preserves the friendly legacy binding
	// behavior and is the zero-value default.
	BreakpointBindNextExecutableInFile BreakpointBindingMode = iota
	BreakpointBindExact
	BreakpointBindNextExecutableInFunction
)

func ReasonFromString(s string) Reason {
	switch s {
	case "entry":
		return ReasonEntry
	case "breakpoint":
		return ReasonBreakpoint
	case "step":
		return ReasonStep
	case "pause":
		return ReasonPause
	case "runtime-error":
		return ReasonRuntimeError
	case "completed":
		return ReasonCompleted
	case "terminated":
		return ReasonTerminated
	default:
		return ""
	}
}

func BreakpointBindingModeFromString(s string) BreakpointBindingMode {
	switch s {
	case "next-executable-in-file":
		return BreakpointBindNextExecutableInFile
	case "exact":
		return BreakpointBindExact
	case "next-executable-in-function":
		return BreakpointBindNextExecutableInFunction
	default:
		return BreakpointBindNextExecutableInFile
	}
}
