# Universal Ferret API

Portable contracts for compiled execution and retained source-level debugging.
Implementations own engine configuration; consumers share `Runtime`, reusable
`Plan`, per-run `Session`, and the types in `source`, `result`, `debugger`, and
`diagnostics`.

## Execution and ownership

`Runtime.Compile` and `CompileDebug` finish compilation before returning a plan.
Syntax and compiler errors are immediate. Plans support repeated and concurrent
sessions with independent parameters and filesystem configuration.
`Plan.Params(ctx) ([]string, error)` returns a caller-owned snapshot of parameter
names or a metadata retrieval error. An empty list with a nil error means the plan
has no parameters; it is distinct from a retrieval failure.

`Runtime.Version(ctx) (Version, error)` reports the version of the runtime
implementation represented by that `Runtime`. A remote adapter reports the remote
runtime's version. This is separate from the embedding application's, CLI's,
daemon/server's, or transport protocol's version. `Version` is an opaque string-backed
value whose `String()` method preserves the implementation-provided value unchanged.
Values such as `v2.0.0-alpha.55`, `2.0.0`, `dev`, and `unknown` are valid; UAPI does
not parse, normalize, or validate them as semantic versions.

Both metadata operations may fail and may require remote I/O. Local implementations
may return immediately. UAPI imposes no transport-specific behavior or caching
requirements.

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
running work and its output before reusing or closing the session, unless the
implementation explicitly supports a stronger contract. Debugger session closure
terminates and settles active commands.

`Runtime.Run` and `Session.Run` return `(Output, error)`. A successful `Run` obtains
a usable, caller-owned handle; it does not certify execution or delivery success.
Execution starts before the handle is returned and is not deferred until consumption.
Admission/preparation failures return a nil handle and an error, preserving cleanup
failures for resources acquired before failure. Implementations must not return a
usable handle alongside a `Run` error, a typed-nil handle, or `(nil, nil)`.

Once a usable handle is returned, terminal execution, encoding, delivery, and
output-owned cleanup errors are reported through `Consume` or `Collect`. The handle
is usable even when no content is ultimately available. Callers running scripts only
for side effects must still consume output to observe completion. `Close` is abandonment;
its success does not certify successful execution.

`Runtime.Run` owns its temporary session and plan. Resources still needed by output
transfer to that output and are finalized by consumption or closure. Resources already
independent of consumption may be released earlier. `Session.Run` output never closes
the caller-owned session. These rules do not introduce cascading parent closure.

## Consumable output

The `result` package defines `Output`, `Content`, `Metadata`, and `Consumer`, also
exported as root `api` aliases:

```text
Runtime.Run / Session.Run -> Output
                              Consume: receive borrowed encoded chunks
                              Collect: obtain detached *Content
                              Close: abandon and release owned resources
```

`Output.Metadata()` returns immutable, local metadata without I/O, including during
consumption and after closure. `Run` establishes it before exposing the handle without
waiting for or buffering the complete payload solely to determine length. `ContentType`
identifies the encoded representation. When `LengthKnown` is true, `Length` is the
nonnegative, exact total encoded payload byte count, not records, transport frames,
remaining bytes, or progress. Unknown length is normal. Known zero length does not
establish presence, and advertised length must not require unrestricted preallocation.

`Consume` and `Collect` are alternative one-shot terminal operations. Admission validates
a non-nil context, then a non-nil consumer for `Consume`, then already-canceled consumption
and invocation contexts, before atomically claiming the handle. Invalid or rejected calls
do not claim it or disrupt another consumer. A rejected consumption context may be retried
within the invocation's lifetime. Once admitted, an operation cannot be resumed or
repeated; failure or cancellation finalizes its owned resources before return.

Valid competing consumption calls match `result.ErrInUse` while delivery is active and
stopping has not begun. After closure or finalization begins, new valid calls match
`result.ErrClosed`. Root convenience exports `api.ErrOutputInUse` and `api.ErrOutputClosed`
reference those same sentinel values. Use `errors.Is`; wrappers and joined failures are
allowed. There is no separate consumed or finalized error.

`Close` abandons unread output without silently draining arbitrarily large payloads.
During active consumption it requests cancellation and waits for the callback and cleanup.
If explicit closure wins before the terminal outcome is committed, the active operation
matches `ErrClosed`, preserving other observed failures. Closure after commitment waits
for cleanup without changing that outcome. Automatic finalization does not add `ErrClosed`
to the original result; caller-context cancellation retains its context error. Repeated
or concurrent `Close` calls return the recorded cleanup outcome, usually nil, rather than
a lifecycle error simply because the output was already closed.

Callbacks are synchronous, ordered, and non-overlapping. No callback remains active or
is invoked after `Consume` returns. Chunks are borrowed, read-only byte slices valid only
during the callback; copy retained bytes. Boundaries need not align with JSON values,
lines, records, or characters. Empty chunks are not EOF. A callback error stops further
delivery and initiates finalization. Panic unwinding also finalizes resources without
swallowing the panic; cleanup failures remain available through `Close`.

Metadata reads are safe concurrently and from callbacks. Reentrant consumption is
rejected under the same admission rules. A callback must not call `Close` synchronously
on its own output because `Close` waits for that callback. Cancellation cannot forcibly
interrupt arbitrary callback code or release borrowed buffers still in use.

The invocation context bounds the output's lifetime after `Run` returns. A consumption
context may shorten that lifetime, not extend or revive it. The effective context passed
to callbacks observes cancellation from either context and their earliest deadline.
Keep any locally created invocation context alive until output is settled; do not defer
its cancellation in a helper that returns a live handle.

### Content presence and errors

`Content` is detached, recipient-owned data that survives output closure. Inspect available
content independently of the collection error:

| Situation | `Collect` | `Consume` |
| --- | --- | --- |
| Absent content | nil content | No callback |
| Present empty content | Non-nil `*Content`, even with nil `Data` | At least one empty chunk |
| Available content plus terminal error | Content and error together | Delivered bytes and terminal error |
| Failure after receiving a prefix | Prefix and error together | Delivered prefix and error |

Neither zero length, nil `Data`, nor empty content type is an absence sentinel.
`Content.Metadata` preserves the complete-payload descriptor even after partial failure;
`len(Content.Data)` counts bytes actually collected. An otherwise complete delivery or
collection that mismatches a known length must return an error. Joined errors preserve
execution, consumer, and cleanup causes through standard Go error traversal.

Consumable delivery permits bounded buffers and reuse without requiring incremental
query evaluation or native encoding. A buffered adapter may transfer suitable detached
owned bytes directly from `Collect`; a generic chunk accumulator or another copy is not
required. Detached content must not alias borrowed or reusable implementation buffers.

### Collecting

This helper returns available content even when collection fails. Fallback closure also
preserves cleanup failures; `Close` is idempotent after collection. These examples use
only the portable API and have compiling counterparts in `output_example_test.go`.

```go
package example

import (
    "context"
    "errors"
    "fmt"

    "github.com/MontFerret/api"
)

func collectQuery(ctx context.Context, runtime api.Runtime, src api.Source) (content *api.Content, err error) {
    output, err := runtime.Run(ctx, src)
    if err != nil {
        return nil, err
    }
    defer func() { err = errors.Join(err, output.Close()) }()
    return output.Collect(ctx)
}

func inspectQuery(ctx context.Context, runtime api.Runtime, src api.Source) error {
    content, err := collectQuery(ctx, runtime, src)
    if content != nil {
        fmt.Printf("%s: %q\n", content.Metadata.ContentType, content.Data)
    }
    return err
}
```

### Forwarding chunks

The destination handles each borrowed chunk synchronously. Destination errors stop
delivery; a short write without an error becomes `io.ErrShortWrite`. This helper retains
no payload. A destination may choose to buffer it, write a file, or forward it elsewhere.

```go
package example

import (
    "context"
    "errors"
    "io"

    "github.com/MontFerret/api"
)

func streamQuery(ctx context.Context, runtime api.Runtime, src api.Source, dst io.Writer) (err error) {
    output, err := runtime.Run(ctx, src)
    if err != nil {
        return err
    }
    defer func() { err = errors.Join(err, output.Close()) }()
    return output.Consume(ctx, func(ctx context.Context, chunk []byte) error {
        if err := ctx.Err(); err != nil {
            return err
        }
        n, err := dst.Write(chunk)
        if err != nil {
            return err
        }
        if n != len(chunk) {
            return io.ErrShortWrite
        }
        return nil
    })
}
```

For side-effect-only execution, consume with a callback that returns nil without
retaining bytes. Observe the consumption error; merely closing the handle abandons it.

Non-nil caller contexts control cancellation of `Run`, `Compile`,
`CompileDebug`, `NewSession`, `NewDebugSession`, `Plan.Params`, and `Runtime.Version`.
All of these operations require a non-nil context. Cancellation errors
preserve `context.Canceled` and `context.DeadlineExceeded` through `errors.Is`.
Parent Close does not require deriving operation contexts to coordinate descendants.
Implementations may use internal contexts for their own resources and may translate portable
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
`WithOutputContentType` and `SetOutputContentType` retain their names. The selected
representation is reported by `Output.Metadata().ContentType`; encoding failures after
a usable handle is returned are observed through `Consume` or `Collect`.

`WithOptimizationLevel` rejects values outside the portable enum during callback
application. Each runtime defines which known optimization levels it supports
and any restrictions for debug compilation.

## Portable data

Native Ferret produces one-based lines and byte columns, with zero-based,
half-open byte spans. Source names are identities and need not be filesystem
paths; anonymous sources have an empty name. Adapters translate native indexed
source text into the portable `Source` representation. Portable coordinates, debugger
values, variables, frames, breakpoints, and reasons retain their existing representations.
Encoded content uses the new nested metadata shape described below. Debugger events
retain their field names, including `output`, whose value is materialized `*Content`.

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

A completed event and its detached content can accompany a later cleanup error.
Retained events and snapshots must not store live output handles or mutable implementation
buffers; their bytes remain readable after debugger cleanup. Inspecting one observer's
event cannot consume another observer's data. Recipients coordinate mutations of shared
materialized content themselves. Error projections should preserve each diagnostic's
source, annotation order, joined branches, and native causes through standard Go error
traversal.
`Event.Error` remains a Go error; this change defines no portable JSON error codec.
Adapters continue to own serialized error projections.

## Migration from materialized output

This is an intentional breaking change to both the Go API and encoded-content JSON:

| Previous API | Current API |
| --- | --- |
| Materialized `result.Output` / `api.Output` struct | Detached `result.Content` / `api.Content` struct |
| `Run(...) (*Output, error)` | `Run(...) (Output, error)`, followed by `Consume` or `Collect` |
| `output.ContentType` | `output.Metadata().ContentType` or `content.Metadata.ContentType` |
| Materialized payload field `Content` | `Content.Data` |
| Flat `contentType` and `content` JSON keys | Nested `metadata` object and `data` key |
| Execution/cleanup result observed at `Run` return | Handle admission at `Run`; terminal result through consumption |
| Debugger event `Output` containing old output struct | Same field/key containing detached `*Content` |

For example, populated materialized content serializes as:

```json
{"metadata":{"contentType":"application/json","length":2,"lengthKnown":true},"data":"NDI="}
```

Payload bytes use standard base64 byte-slice encoding; they are not interpreted as a
JSON document. A nil `*Content` serializes as `null`. A present content object with nil
`Data` has `"data":null`; a non-nil empty byte slice has `"data":""`. All descriptor fields
remain present, including unknown length. Old JSON keys are not emitted. Live output
handles must not be serialized, and marshaling must never consume them.
