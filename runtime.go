package api

import (
	"context"
	"io"
)

// Runtime compiles source into reusable plans. Callers own directly created
// plans and sessions and close them from children to parents before the runtime.
// Close is idempotent, retains its cleanup result, and rejects new work.
// Admitted creation settles before parent cleanup. Descendant cleanup is not guaranteed.
// Contexts must be non-nil. Run closes its temporary session and plan, preserving
// execution and cleanup errors together with any available encoded output.
type Runtime interface {
	io.Closer
	Run(ctx context.Context, src Source, opts ...SessionOption) (Output, error)
	Compile(ctx context.Context, src Source, opts ...PlanOption) (Plan, error)
	CompileDebug(ctx context.Context, src Source, opts ...PlanOption) (Plan, error)
}
