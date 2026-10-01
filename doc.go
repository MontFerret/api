// Package api defines implementation-independent contracts for compiling,
// executing, and debugging Ferret queries.
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
