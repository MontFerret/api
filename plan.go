package api

import (
	"context"
	"io"

	"github.com/MontFerret/api/debugger"
)

type (
	// Plan represents a compiled program. Compilation finishes before a plan is
	// returned. Plans support independent sessions and are not consumed by execution.
	// Params returns a caller-owned snapshot or an error if metadata cannot be retrieved.
	//
	// Close releases plan-owned resources and prevents subsequent session and
	// debug-session creation. It is idempotent and retains its cleanup result,
	// without requiring identical error-wrapper pointers. Close need not wait for
	// constructors already started. It does not implicitly close or cancel returned
	// sessions or debug sessions; callers remain responsible for their lifecycle.
	//
	// NewSession and NewDebugSession use non-nil caller contexts for cancellation.
	// Close does not cancel those contexts. Callers coordinate work and cleanup
	// when sessions use plan-owned resources.
	Plan interface {
		io.Closer
		Params() ([]string, error)
		NewSession(ctx context.Context, opts ...SessionOption) (Session, error)
		NewDebugSession(ctx context.Context, opts ...SessionOption) (debugger.Session, error)
	}
)
