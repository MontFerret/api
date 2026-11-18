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
// error-wrapper pointers. Inspection references expire on resume or termination.
// A command can return an event and an error, including available completion
// content when subsequent cleanup fails. Event.Output holds detached materialized
// Content that remains readable after cleanup; inspecting a retained event never
// consumes a live output handle. Context arguments must be non-nil.
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

	// Variables reads one bounded page of immediate children in the current
	// paused state. Page arguments follow VariablePage.Normalize. Select the
	// child kind before applying the range. An absent kind or a start at or
	// beyond the end returns an empty page. A short successful page means the
	// selected sequence is exhausted, including when its length is unknown.
	// Recognized modes that cannot be paged safely return an error matching
	// errors.ErrUnsupported. Invalid, unknown, or expired references while
	// paused return an error matching ErrInvalidValueReference.
	//
	// Callers may retain the page as a read-only snapshot. Its slice entries
	// and counts remain readable across later calls, resume, and Close. Only
	// the expansion references expire.
	// Repeated reads of the same position in a parent's named or indexed
	// sequence reuse the child's reference for that stop.
	//
	// The producing session checks the context during reads and before publishing
	// the complete page and its new handles together. A malformed UTF-8 key in
	// the requested page is a child-read error. Read failure, cancellation,
	// termination, or handle exhaustion before publication returns a nil page
	// and an error. Exhaustion preserves previously issued handles. If remote
	// delivery fails after publication, the caller also receives a nil page
	// and an error. Context errors preserve errors.Is identity. Canceling an
	// inspection affects only that request.
	Variables(ctx context.Context, reference ValueReference, page VariablePage) ([]Variable, error)
	Evaluate(ctx context.Context, expression string) (Value, error)
	EvaluateFrame(ctx context.Context, frame int, expression string) (Value, error)
}
