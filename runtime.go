package api

import (
	"context"
	"io"
)

// Runtime compiles source into reusable plans. Callers own directly created
// plans and sessions and are responsible for their cleanup.
//
// Close releases implementation-owned resources and is idempotent, retaining
// its cleanup result without requiring identical error-wrapper pointers.
// Owning runtimes reject subsequent work according to their closed-state
// semantics. Borrowing adapters may document a no-op Close that leaves the
// adapter and underlying runtime usable. Close does not implicitly cancel
// caller-owned work and need not wait for operations already started.
//
// Run, Compile, CompileDebug, and Version use non-nil caller contexts for cancellation.
// Callers coordinate work and cleanup when descendants use parent-owned
// resources. Run closes its temporary session and plan, preserving execution
// and cleanup errors together with any available encoded output.
type Runtime interface {
	io.Closer

	// Version reports the version of the runtime implementation represented by
	// this Runtime. A remote adapter reports its remote runtime's version, not
	// the version of the host application, CLI, daemon/server, or transport protocol.
	//
	// ctx must be non-nil. Cancellation errors must preserve context.Canceled
	// and context.DeadlineExceeded through errors.Is. Retrieval may involve
	// remote I/O; local implementations may return immediately. UAPI imposes
	// no transport-specific behavior or caching requirements.
	Version(ctx context.Context) (Version, error)

	// Run returns nil output with an error when no output was produced.
	// A non-nil output with a nil error indicates success, including empty output.
	// A non-nil output may accompany an error from cleanup or other processing;
	// callers must inspect output independently of the error.
	Run(ctx context.Context, src Source, opts ...SessionOption) (*Output, error)
	Compile(ctx context.Context, src Source, opts ...PlanOption) (Plan, error)
	CompileDebug(ctx context.Context, src Source, opts ...PlanOption) (Plan, error)
}
