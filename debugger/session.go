package debugger

import (
	"context"
	"io"

	"github.com/MontFerret/api/source"
)

// Session controls one retained execution. Start establishes its lifetime;
// resume commands observe both that lifetime and their non-nil caller context.
// Commands are serialized. Pause and Close may interrupt an active command.
// Close terminates execution, waits for commands, and releases resources;
// repeated closes retain the cleanup result without requiring identical
// error-wrapper pointers. Inspection references expire on resume.
// A command can return an event and an error, including available completion
// output when subsequent cleanup fails. Context arguments must be non-nil.
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
