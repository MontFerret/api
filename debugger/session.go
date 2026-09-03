package debugger

import (
	"context"
	"io"

	"github.com/MontFerret/api/source"
)

type Session interface {
	io.Closer
	Start(ctx context.Context) (*Event, error)
	Continue(ctx context.Context) (*Event, error)
	StepIn(ctx context.Context) (*Event, error)
	StepOver(ctx context.Context) (*Event, error)
	StepOut(ctx context.Context) (*Event, error)
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
