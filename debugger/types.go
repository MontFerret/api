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
		Type      string
		Display   string
		Reference ValueReference
	}

	// Variable describes a visible local or bind parameter.
	Variable struct {
		Name    string
		Value   Value
		Mutable bool
		Param   bool
	}

	// Frame describes the paused top frame or one of its callers.
	Frame struct {
		Name       string
		Location   source.Location
		FunctionID FunctionID
	}

	// Breakpoint describes a requested source-location breakpoint and its resolved
	// executable location, when one exists.
	Breakpoint struct {
		Location          source.Location
		RequestedPosition source.Position
		ID                BreakpointID
		PointID           PointID
		FunctionID        FunctionID
		BindingMode       BreakpointBindingMode
		Bound             bool
	}

	// BreakpointOptions configures how a requested source location binds.
	BreakpointOptions struct {
		BindingMode BreakpointBindingMode
	}

	// Event reports a debugger stop, completion, or termination.
	Event struct {
		Error            error
		Output           *result.Output
		Reason           Reason
		HitBreakpointIDs []BreakpointID
		Location         source.Location
		Depth            int
	}

	Session interface {
		io.Closer

		Start(ctx context.Context) (*Event, error)
		Continue(ctx context.Context) (*Event, error)
		Step(ctx context.Context) (*Event, error)
		Next(ctx context.Context) (*Event, error)
		Out(ctx context.Context) (*Event, error)
		Pause() error
		SetBreakpoint(file string, line int) (Breakpoint, error)
		SetBreakpointAt(location source.Location, opts BreakpointOptions)
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
