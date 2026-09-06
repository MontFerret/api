package api

import (
	"context"
	"io"

	"github.com/MontFerret/api/debugger"
)

type (
	// Plan represents an already compiled program. Compilation errors are returned
	// before a plan is published. Plans support independent sessions and are not
	// consumed by execution. Callers close sessions before their plan. Close is
	// idempotent, retains its cleanup result, and prevents new sessions. Admitted
	// creation settles before parent cleanup; descendant cleanup is not guaranteed.
	// Params returns a caller-owned snapshot. Contexts must be non-nil.
	Plan interface {
		io.Closer
		Params() []string
		NewSession(ctx context.Context, opts ...SessionOption) (Session, error)
		NewDebugSession(ctx context.Context, opts ...SessionOption) (debugger.Session, error)
	}
)
