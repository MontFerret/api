package api

import (
	"context"
	"io"
)

// Session executes a compiled plan with per-session configuration. Run observes
// its non-nil context and returns a caller-owned consumable Output. Unless a
// stronger implementation contract explicitly allows otherwise, callers serialize
// Run and settle its output before reusing or closing the session. Output cleanup
// never closes this caller-owned session. Close is idempotent and retains its
// cleanup result, without requiring identical error-wrapper pointers.
type Session interface {
	io.Closer

	// Run starts execution and returns a usable Output with nil error, rather than
	// certifying execution or delivery success. Execution is not deferred until
	// consumption. Metadata is reliable before return without buffering solely
	// to determine length.
	//
	// Admission or preparation failures return nil output and an error, preserving
	// cleanup failures for resources acquired by this invocation. Never return a
	// usable handle alongside an error or represent absence using a typed nil.
	// After a handle is returned, terminal execution, encoding, delivery, and
	// output-owned cleanup errors belong to Consume or Collect, alongside any
	// available content. An absent payload still has a usable output handle.
	//
	// ctx must be non-nil and bounds the output's lifetime after return. Keep it
	// alive until the output is settled. Consume or Collect observes completion
	// and finalizes output-owned resources without closing this session. Close on
	// the output abandons it; side-effect-only callers must consume. See Output.
	Run(ctx context.Context) (Output, error)
}
