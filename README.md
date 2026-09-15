# Universal Ferret API

Portable contracts for compiled execution and retained source-level debugging.
Implementations own engine configuration; consumers share `Runtime`, reusable
`Plan`, per-run `Session`, and the types in `source`, `result`, `debugger`, and
`diagnostics`.

## Execution and ownership

`Runtime.Compile` and `CompileDebug` finish compilation before returning a plan.
Syntax and compiler errors are immediate. Plans support repeated and concurrent
sessions with independent parameters and filesystem configuration.
`Params() ([]string, error)` returns a caller-owned snapshot or a metadata
retrieval error. It does not take a context.

Each object releases the resources it owns. Owning runtimes reject subsequent
work according to their closed-state semantics. Borrowing adapters may document
a no-op `Close` that leaves the adapter and underlying runtime usable; the
underlying owner remains responsible for cleanup.

Plan closure prevents subsequent session and debug-session creation. It does
not implicitly close or cancel already-created sessions or debug sessions.
Returned descendants remain responsible for their own lifecycle. Parent closure
does not implicitly cancel caller-owned operations and need not wait for
operations or constructors already started. Callers coordinate outstanding work
and cleanup when descendants use parent-owned resources.

Close is idempotent and retains the completed cleanup result, without requiring
identical error-wrapper pointers. Ordinary execution owners cancel and settle
running work before closing the session. Debugger session closure terminates
and settles active commands. `Runtime.Run` owns its temporary session and plan
and returns execution and cleanup errors together, preserving available output.
Output and inspection snapshots belong to their callers.

`Runtime.Run` and `Session.Run` return `(*Output, error)`. Output presence is
independent of the error:

| Output | Error | Meaning |
| --- | --- | --- |
| nil | non-nil | No output was produced. |
| non-nil | nil | Execution succeeded, including empty output. |
| non-nil | non-nil | Output was produced, but cleanup or other processing also failed. |

A non-nil `&Output{}` is present output; the zero value is not an absence
sentinel. Inspect output independently of the error:

```go
output, err := runtime.Run(ctx, api.NewAnonymousSource("RETURN 42"))
if output != nil {
    consume(output.ContentType, output.Content)
}
if err != nil {
    return err
}
```

The output fields and their serialized representation are unchanged. Transports
preserve output presence through their own representations.

Non-nil caller contexts control cancellation of `Run`, `Compile`,
`CompileDebug`, `NewSession`, and `NewDebugSession`. Cancellation errors
preserve `context.Canceled` and `context.DeadlineExceeded` through `errors.Is`.
Implementations need not derive operation contexts to coordinate parent Close.
They may use internal contexts for their own resources and may translate portable
option callbacks before validating the operation context.

## Options

Session options target an implementation of `SessionOptions`, which may apply
settings directly or queue implementation-specific options. Non-nil callbacks
run exactly once in order, and their returned errors are joined. Later options
can override earlier values; parameter maps merge. Runtime-specific extensions
validate their target explicitly.

Portable or application-level errors may be returned immediately by an option
callback. Runtime-specific conversion and validation may be deferred until the
setting is used. This includes host-parameter conversion, output codecs,
filesystem-root construction, and runtime-specific capabilities. A nil setter
error therefore does not certify runtime validity.

Invalid settings must fail the operation no later than their relevant point of use.
Implementations need not preflight all session configuration before compilation,
resource acquisition, or execution. `Runtime.Run` may compile, create a session,
then execute. Output codec availability may be validated during result encoding,
after the query has run. Implementations document validation timing and when
mutable inputs are converted or snapshotted.

`WithOptimizationLevel` rejects values outside the portable enum during callback
application. Each runtime defines which known optimization levels it supports
and any restrictions for debug compilation.

## Portable data

Native Ferret produces one-based lines and byte columns, with zero-based,
half-open byte spans. Source names are identities and need not be filesystem
paths; anonymous sources have an empty name. Adapters translate native indexed
source text into the portable `Source` representation. Portable coordinates,
encoded output, debugger values, variables, frames, breakpoints, reasons, and
events preserve their existing fields and JSON representations.

`debugger.ValueReference.Valid` accepts positive references. References are scoped
to a paused state and become stale when execution resumes. `debugger.NoFunction`
identifies the top-level program body. Compiler table identifiers and validation
remain implementation details.

Debugger commands are `StepIn`, `StepOver`, and `StepOut`. Breakpoint requests use
canonical source locations and binding options. `ReplaceBreakpoints(ctx, sourceName,
requests)` replaces a source's complete set atomically, including during execution;
an empty request slice clears it. Each `BreakpointRequest` contains a position and
binding options. Results preserve request order and distinguish unbound locations
from operation failure. Failure before publication leaves the prior set intact.
A stop already decided before removal remains inspectable with its original hit
IDs. Incremental add/delete methods remain available and observe their request
context through publication; canceling a breakpoint request does not cancel the
debuggee. All debugger methods except Close take a non-nil context, including
Pause, breakpoint listing, and inspection. Inspection checks cancellation before
and after command admission; cancellation need not interrupt the admission wait.
A canceled pause request must not request a stop.

`Breakpoints(ctx) ([]Breakpoint, error)` returns a detached, ID-ordered snapshot,
including after Close. Nil and canceled contexts return errors. Metadata and
listing errors can be reported without conflating failure with an empty result.

A completed event and its output can accompany a later cleanup error. Error projections should preserve each
diagnostic's source, annotation order, joined branches, and native causes through
standard Go error traversal.
