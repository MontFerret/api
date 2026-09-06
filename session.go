package api

import (
	"context"
	"io"
)

// Session executes a compiled plan with per-session configuration. Run observes
// its non-nil context and returns caller-owned encoded output. Unless documented
// otherwise, callers serialize Run and settle it before Close. Close is
// idempotent and retains its cleanup result.
type Session interface {
	io.Closer
	Run(c context.Context) (Output, error)
}
