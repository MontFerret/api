// Package api defines implementation-independent contracts for compiling,
// executing, and debugging Ferret queries.
//
// Runtime.Run and Session.Run return caller-owned, one-shot Output handles.
// A successful Run obtains a usable handle; Consume or Collect observes terminal
// execution, delivery, and cleanup errors. Close abandons unread output. Invocation
// contexts remain in effect until output is settled. Output.Metadata is immutable
// and local; Content is detached materialized data, also used by debugger events.
// See result.Output for consumption, presence, cancellation, and closure contracts.
//
// Plan.Params and Runtime.Version retrieve metadata and may involve remote I/O.
// Both require non-nil caller contexts, and cancellation errors must preserve
// context.Canceled and context.DeadlineExceeded through errors.Is. The API imposes
// no transport-specific behavior or caching requirements.
//
// Plan.Params returns a caller-owned snapshot; an empty parameter list is distinct
// from a metadata retrieval error. Runtime.Version returns an opaque Version for
// the represented runtime implementation, including the remote runtime for a
// remote adapter, rather than the host application, CLI, server, or transport.
package api
