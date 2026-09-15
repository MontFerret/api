package debugger

import (
	"context"
	"io"

	"github.com/MontFerret/api/source"
)

// Session controls one retained execution. Start establishes its lifetime;
// resume commands observe both that lifetime and their non-nil caller context.
// Execution and inspection commands are serialized. Breakpoint operations may
// run concurrently with execution. Pause and Close may interrupt an active command.
// Close terminates execution, waits for commands, and releases resources;
// repeated closes retain the cleanup result without requiring identical
// error-wrapper pointers. Inspection references expire on resume.
// A command can return an event and an error, including available completion
// output when subsequent cleanup fails. Context arguments must be non-nil.
// Inspection checks cancellation before and after command admission; cancellation
// need not interrupt the admission wait. A canceled Pause must not request a stop.
// Breakpoint mutations observe their request context through publication; canceling
// a request does not cancel the execution. Cancellation observed before publication
// aborts the mutation; cancellation after publication does not undo success.
type Session interface {
	io.Closer
	Start(ctx context.Context) (*Event, error)
	Continue(ctx context.Context) (*Event, error)
	StepIn(ctx context.Context) (*Event, error)
	StepOver(ctx context.Context) (*Event, error)
	StepOut(ctx context.Context) (*Event, error)
	Pause(ctx context.Context) error

	// ReplaceBreakpoints atomically replaces one source's complete requested set,
	// before execution, while paused, or while running. Empty requests clear it;
	// an empty sourceName selects the launched source. Other sources are unchanged.
	// Results follow request order. Invalid positions or binding modes fail the
	// operation without publication; valid but unresolved locations return Bound=false.
	//
	// Unchanged requests retain IDs; duplicates match in ascending ID order.
	// Removed IDs are never reused in the session.
	// A hit already decided against an older set retains its IDs and remains an
	// inspectable stop. Additions affect subsequent visits, never past instructions.
	//
	// Cancellation observed before publication leaves the previous set intact.
	// After publication the operation returns success even if cancellation follows.
	// Session terminal commitment and publication are ordered: publication first
	// succeeds; completion, termination, or Close first rejects replacement.
	// Concurrent writers are serialized in admission order. Callers requiring
	// request order must await each result before issuing the next replacement.
	ReplaceBreakpoints(ctx context.Context, sourceName string, requests []BreakpointRequest) ([]Breakpoint, error)
	SetBreakpoint(ctx context.Context, pos source.Location) (Breakpoint, error)
	SetBreakpointAt(ctx context.Context, loc source.Location, opts BreakpointOptions) (Breakpoint, error)
	DeleteBreakpoint(ctx context.Context, id BreakpointID) error

	// Breakpoints returns a detached, ID-ordered snapshot, including after Close.
	// Nil or canceled contexts return an error.
	Breakpoints(ctx context.Context) ([]Breakpoint, error)
	Frames(ctx context.Context) ([]Frame, error)
	Locals(ctx context.Context) ([]Variable, error)
	FrameLocals(ctx context.Context, frame int) ([]Variable, error)
	Variables(ctx context.Context, reference ValueReference) ([]Variable, error)
	Evaluate(ctx context.Context, expression string) (Value, error)
	EvaluateFrame(ctx context.Context, frame int, expression string) (Value, error)
}
