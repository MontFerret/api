package api

import (
	"context"
	"io"

	"github.com/MontFerret/api/debugger"
)

type (
	// Plan represents a compiled program. Compilation finishes before a plan is
	// returned. Plans support independent sessions and are not consumed by execution.
	//
	// Close releases plan-owned resources and prevents subsequent session and
	// debug-session creation. It is idempotent and retains its cleanup result,
	// without requiring identical error-wrapper pointers. Close need not wait for
	// constructors already started. It does not implicitly close or cancel returned
	// sessions or debug sessions; callers remain responsible for their lifecycle.
	//
	// Params, NewSession, and NewDebugSession use non-nil caller contexts for cancellation.
	// Close does not cancel those contexts. Callers coordinate work and cleanup
	// when sessions use plan-owned resources.
	Plan interface {
		io.Closer

		// Params returns a caller-owned snapshot of the plan's parameter names.
		// An empty list with a nil error means the plan has no parameters; failure
		// to retrieve metadata returns an error.
		//
		// ctx must be non-nil. Cancellation errors must preserve context.Canceled
		// and context.DeadlineExceeded through errors.Is. Retrieval may involve
		// remote I/O; local implementations may return immediately. UAPI imposes
		// no transport-specific behavior or caching requirements.
		Params(ctx context.Context) ([]string, error)
		NewSession(ctx context.Context, opts ...SessionOption) (Session, error)
		NewDebugSession(ctx context.Context, opts ...SessionOption) (debugger.Session, error)
	}
)
