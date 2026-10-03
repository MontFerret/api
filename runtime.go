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
// resources. Run owns its temporary session and plan; resources needed for
// consumption transfer to the returned Output, which finalizes them. Resources
// independent of consumption may be released earlier. Caller-owned descendants
// and borrowed parents retain their existing ownership.
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

	// Run starts execution and returns a usable, caller-owned Output with nil
	// error. Success means a handle was obtained, not that execution or delivery
	// completed. Execution is not deferred until consumption. Metadata is reliable
	// before return without buffering solely to determine length.
	//
	// Admission or preparation failures return nil output and an error, preserving
	// cleanup failures for resources already acquired. Never return a usable handle
	// alongside a Run error or use a typed-nil implementation as an absent handle.
	// Once a handle is returned, terminal execution, encoding, delivery, and
	// output-owned cleanup errors are reported by Consume or Collect, preserving
	// available content. Even absent content is observed through a usable handle.
	//
	// ctx must be non-nil and bounds the output's lifetime after return. Callers
	// must not cancel it before settling the output. Consume or Collect finalizes
	// resources; Close abandons unread output without certifying completion.
	// Side-effect-only callers must consume to observe completion. See Output.
	Run(ctx context.Context, src Source, opts ...SessionOption) (Output, error)
	Compile(ctx context.Context, src Source, opts ...PlanOption) (Plan, error)
	CompileDebug(ctx context.Context, src Source, opts ...PlanOption) (Plan, error)
}
