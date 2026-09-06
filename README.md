# Universal Ferret API

Portable contracts for compiled execution and retained source-level debugging.
Implementations own engine configuration; consumers share `Runtime`, reusable
`Plan`, per-run `Session`, and the types in `source`, `result`, `debugger`, and
`diagnostics`.

## Execution and ownership

`Runtime.Compile` and `CompileDebug` finish compilation before returning a plan.
Syntax and compiler errors are immediate. Plans support repeated and concurrent
sessions with independent parameters and filesystem configuration; `Params`
returns a caller-owned snapshot.

Callers close directly created sessions before plans, then close the runtime.
Parents reject new children after closure begins and coordinate admitted
construction with closure; they do not promise to close descendants. Close is
idempotent and retains the completed cleanup result. Ordinary execution owners
cancel and settle running work before closing the session. Debugger session
closure terminates and settles active commands. `Runtime.Run` owns its temporary
session and plan and returns execution and cleanup errors together, preserving
available output. Output and inspection snapshots belong to their callers.

Contexts must be non-nil. Already-canceled contexts are rejected before options,
hooks, or acquisition. Implementations recheck cancellation before publishing
resources and preserve `context.Canceled` and `context.DeadlineExceeded` through
`errors.Is`. Synchronous native compilation checks cancellation between phases;
it does not preempt the parser or detach compilation into a goroutine.

## Options

Session options target the implementation's actual `SessionOptions` owner.
Non-nil options run exactly once in order; their validation failures are joined
before acquiring resources. Later options can override earlier values. Native
extensions validate their target explicitly. No intermediate option replay is
required.

`WithOptimizationLevel` configures one compilation. Native Ferret inherits the
engine's level when omitted, supports explicit `OptimizationNone`,
`OptimizationBasic`, and `OptimizationFull` without mutating shared compiler
configuration, and rejects `OptimizationAggressive`. Debug compilation accepts
omission or `OptimizationNone` only. Other implementations document their
supported levels.

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
canonical source locations and binding options. A completed event and its output
can accompany a later cleanup error. Error projections should preserve each
diagnostic's source, annotation order, joined branches, and native causes through
standard Go error traversal.
