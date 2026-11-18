# Debugger value inspection

A paused session describes a value without reading its children. A caller
expands that value by requesting one bounded page. The Universal API owns this
contract. Core owns safe access to runtime values, Wire carries the same
operation, and clients decide when to expand a node.

## Value metadata

`debugger.Value` contains `Type string`, `Display string`,
`Reference ValueReference`, and optional `IndexedCount, NamedCount *int64`.
Locals, evaluation results and children all use this description.

- `Type` is a semantic FQL or resource type. `Display` is a cheap summary,
  such as `Array(1000000)`, `Object(42)` or `Binary(1024)`. An opaque resource
  can supply a name such as `SQL::Rows`. Fallbacks must keep concrete Go types,
  package paths and pointer strings private.
  Both fields are valid UTF-8. The presentation budget bounds the work and
  memory used to produce them, including validation, escaping and truncation.
- Counts are read-only, detached metadata snapshots. A nonnegative count gives
  the exact number of children of that kind. Zero means known empty, and nil
  means unknown. Counts may be omitted independently. A binary's byte length
  in its display, for example, need not be a child count.
- A positive reference permits safe expansion in at least one mode. Zero means
  there is no expansion. A known empty array has indexed count zero and
  reference zero. A summary-only object can report a positive known size.
  A safely pageable collection keeps its reference regardless of its size.

## Pages

```go
type VariableFilter string

type VariablePage struct {
	Filter VariableFilter
	Start  int64
	Count  int64
}

// Session method in package debugger
Variables(ctx context.Context, reference ValueReference, page VariablePage) ([]Variable, error)
```

`VariablePage.Normalize() (VariablePage, error)` is the shared source of range
validation. `VariableFilterNamed` (`"named"`) selects named children and
`VariableFilterIndexed` (`"indexed"`) selects indexed children.
`VariableFilterAll` (the empty string) concatenates named then indexed children.
Apply the offset to the selected sequence. An absent kind on a valid reference
returns an empty page.

`Start` is zero-based. An omitted or zero `Count` selects
`DefaultVariablePageSize` (100). The maximum is `MaxVariablePageSize` (1000).
An unknown filter, negative offset or count, count above the maximum, or
overflow of the exclusive end `Start + Count` after defaulting fails with
`ErrInvalidVariablePage`. Oversized requests must be split explicitly.

A successful page contains the requested number of children or all that remain.
A short page establishes the end even when the count is unknown. An offset at
or beyond the end succeeds with zero children for a valid range. For a full
page, advance to its exclusive end and limit the next count to the representable
range. At `math.MaxInt64` no further positive count is representable. A mixed
collection may still have indexed children reachable using the indexed filter.
When the named prefix length is unknown, clients that require lossless
traversal can request separate filtered views from the start.
Reading a page never enumerates an unbounded prefix to reach the offset.

Indexed names are the decimal, zero-based **source** indices. Named children
keep their exact UTF-8 keys in deterministic order throughout the stop. A key
with malformed UTF-8 fails the requested page with an ordinary child-read
error. A literal U+FFFD is a valid key. Validate keys only on the requested
page so a malformed key elsewhere leaves other addressable pages readable.
Preserve the key bytes across JSON and protobuf. Substitution, normalization
and truncation can merge distinct children.

Repeated and overlapping pages agree on names, order and values. To obtain an
indexed child's source position, subtract the named prefix length from its
offset in the combined sequence.
Adding both counts can overflow even for a representable range. The named key
`"0"` and indexed child zero occupy distinct positions.

A recognized mode whose order or boundary cannot be determined safely returns
an error matching `errors.ErrUnsupported`. A positive reference may support
only the named or indexed view. If a `VariableFilterAll` request fails, check
`errors.Is(err, errors.ErrUnsupported)`. On a match, try `VariableFilterNamed`
or `VariableFilterIndexed` on the same reference. Start the new view at offset
zero, even if its count is unknown, and track its pages separately from any
combined results. A short or empty page ends only that kind. Propagate other
errors.

For example, suppose `results` is a local returned by `FrameLocals`. Its value
has type `Array`, display `Array(1000000)`, indexed count 1,000,000, named
count zero, and a positive reference. Describing it reads zero elements. This
request retrieves elements 1000 through 1019.

```go
children, err := session.Variables(ctx, results.Value.Reference, debugger.VariablePage{
	Filter: debugger.VariableFilterIndexed,
	Start:  1000,
	Count:  20,
})
if err != nil {
	return err
}
fmt.Println(children[0].Name, children[19].Name) // 1000 1019
```

Each child is itself metadata. Expanding a nested value requires another
request. A caller can retain a page as a read-only snapshot. Its entries, names,
summaries and counts remain readable and unchanged by later requests, resume
and close. Only the expansion handles expire.

## Safety and ownership

Core can read stable native arrays and VM-owned buffered datasets directly.
The paused session serializes inspection with resume. A host lending mutable
values must also keep their reachable storage stable for the stop, including
borrowed backing slices, or supply an existing stable inspection view. Without
that guarantee return reference zero and only independently safe metadata. Omit
counts that would touch unstable storage. Obtaining a stable view must respect
the same work bounds.

The existing `runtime.DebugInspectable.DebugInfo` hook supplies presentation
only. A custom value can separately opt into debugger-safe access by
guaranteeing cheap metadata, stable order and bounded reads at arbitrary
offsets. Implementing `runtime.List`, `runtime.Map` or `runtime.Queryable`
alone makes no such guarantee.

Metadata and page reads must leave the debuggee unchanged. They must never
run a query, consume an iterator, materialize a stream, perform external I/O,
invoke a user function or call an untrusted formatter or accessor. A cursor or
deferred query without a safe view gets a summary and reference zero.
`Evaluate` and `EvaluateFrame` remain explicit operations with their own
expression rules.

Every reference belongs to its issuing session and paused state. Starting or
resuming execution, a replacement stop, termination or close invalidates the
old references. Never reuse an expired ID within the session.
`ValueReference.Valid()` checks the integer encoding only. During a live stop,
a zero, negative, unknown or expired ID fails with `ErrInvalidValueReference`.
Callers must keep session identity with each ID. Outside a paused state,
`Variables` returns a session-state error.

Within a stop, repeated frame listings reuse the root handle for each
unchanged expandable local or bound parameter in that frame. `Locals(ctx)`
and `FrameLocals(ctx, 0)` list the same bindings and share their references.
The guarantee applies to each binding in its own frame, including recursive
calls with identically named locals.

Reuse the child handle for the same parent reference, kind and position
across repeated or overlapping reads in the stop. Equivalent values under
different parents may have different handles. If an implementation exhausts
its handle budget or positive ID space, fail the operation atomically with a
resource error. Existing handles stay usable until the stop ends.

The session orders concurrent reads and resume with its existing command
serialization. Check the non-nil context before and after admission, during
reads and before publication. Waiting for admission need not be interruptible.
Commit a page and its new handles together, ordered against termination and
close. A child error, cancellation, deadline or termination before commit
returns nil and an error. Request cancellation affects only that request.
Preserve context and inspection sentinel errors for `errors.Is`. Adapters must
check the stop generation before installing handles or updating a client tree.

## Downstream migration

- **Ferret Core.** The current formatter and reference allocator both call
  `DebugValueAccess.Inspect`. Only complete, nonempty snapshots within the
  format limit (eight items by default) get a handle. The VM also falls back
  to a reflected Go type for buffered `data.DataSet`. Separate description,
  reference eligibility and page reads in the VM accessor, debugger and
  `uapi` adapter. Use array semantics and direct indexing for buffered
  datasets. Propagate the request context and child read errors. Format only
  a bounded prefix under Core's `MaxBytes`, including for host presentation
  hints. Today's UTF-8 validation and repair traverse the whole input before
  truncation. A shape with efficient ordered access can page in that order.
  On explicit expansion Core may sort at most 128 unordered keys if both the
  backing scan and key processing are bounded. Sparse or preallocated storage
  can exceed that work even with few live keys. Use a summary for unordered
  maps without a bounded path. The threshold belongs to Core.
- **Wire.** The v1 `Variables` RPC carries only a reference. Change the
  canonical RPC, client, server and value conversions together. Carry the
  filter, signed 64-bit range and presence-aware nonnegative 64-bit counts
  through all debugger value paths. Validate at both ends using the shared
  bounds, checked reference conversions and result-length checks. Preserve
  distinguishable page, reference, unsupported, cancellation and deadline
  errors. Check response text before encoding and bound response bytes
  independently of the item limit. An oversized page fails delivery with a
  resource error and no prefix. The hosted session may already have committed
  its page when size validation, cancellation or transport failure prevents
  delivery. Wire cannot roll back those handles. Retrying the same range within
  the same stop reuses them. Task 4 needs a new Wire protocol revision that
  rejects **both** new-client/old-server and old-client/new-server pairs before
  use. Today the client accepts a nonempty protocol version and the server
  checks an output-framing version.
  A response-only version bump leaves old clients able to connect.
- **ferretd / DAP.** The current adapter rejects paging and filtering. Forward
  the DAP filter and range to the session and apply the same selection to
  named bindings on scope handles. Return the actual child name and semantic
  display. An `evaluateName` must be a valid expression for that variable.
  Emit semantic `type` when the client advertises `supportsVariableType`.
  Reuse stop-scoped DAP handles for the same unchanged occurrence, including
  virtual range nodes. Stage each response's new mappings together so handle
  exhaustion preserves previously published mappings and returns no partial
  response. Expose optional `indexedVariables` and `namedVariables` on
  variables, scopes and evaluation results as applicable.
  Check DAP's unsigned 32-bit ranges and nonnegative signed 32-bit child counts
  and handles. Omit unrepresentable counts and preserve explicit zero where
  reported (go-dap's `int,omitempty` needs a
  presence-aware response). Honor `supportsVariablePaging` from client
  initialization. [DAP specification](https://microsoft.github.io/debug-adapter-protocol/specification#Requests_Variables)
- **DAP requests for all children.** Missing or zero DAP count means all,
  including for clients that support paging. A short successful DAP page
  signals the end. For clients that expand large or unknown groups without a
  range, expose ordinary DAP virtual range or continuation nodes. Choose the
  immediate children when issuing the handle and keep that view stable for
  the stop. Its DAP counts, filters and offsets describe those immediate
  children. Expanding a continuation reads its bounded page. With unknown
  size, a full page can offer a continuation at the next representable offset.
  Its first read may discover the end.
  For mixed values with an unknown named count, an adapter can omit the
  parent's optional counts and expose separate virtual named and indexed
  groups so both are reachable. If a request exceeds the resource budget for
  an already issued handle, fail the request. Silently capping a response
  loses children, and fetching every page defeats the bound.
- **Editorium.** JetBrains currently requests only a reference, discards
  counts and marks the first batch complete. Advertise paging, carry both
  counts through scopes, locals and watches, and request subsequent ranges
  on user demand. The platform's `tooManyChildren` callback accepts `-1`
  for an unknown remainder. Preserve the existing stop-generation and
  obsolete-node checks on every asynchronous result. VS Code forwards DAP
  through the extension. Its native tree pages a positive indexed count,
  fetches named children without a count, and requests all children when the
  indexed count is absent. Exercise both local and Wire sessions with ranges,
  unknown counts and stale replies. Ordinary collection expansion must omit
  the DAP `lazy` evaluation hint. [JetBrains paging API](https://github.com/JetBrains/intellij-community/blob/master/platform/xdebugger-api/src/com/intellij/xdebugger/frame/XCompositeNode.java) · [VS Code variable tree](https://github.com/microsoft/vscode/blob/main/src/vs/workbench/contrib/debug/common/debugModel.ts)
- **CLI.** The current debugger renders `Value.Display` for bindings and
  evaluations. Updating the API and runtime dependencies gives it the same
  semantic summaries without adding a variable tree.

## Downstream acceptance

- **Core.** Record reads from a million-element buffered dataset and confirm
  that description reads none and the example page reads only its 20 elements.
  Exercise a borrowed stable view and a summary-only provider whose backing
  accessors must never run. Put a malformed raw key beside a literal U+FFFD
  key. Verify the malformed key fails its page and the valid key succeeds on
  its own page. Hold one page through another read, resume and close.
  Alternate `Locals` and `FrameLocals(0)`, and repeat a caller-frame listing.
  Verify that refreshing unchanged bindings allocates no additional root handles.
  Re-read overlapping expandable children and verify retained handles plateau.
  Fail or cancel after staging the first new child, exercise close winning
  publication, and exhaust the handle budget and positive ID range. Check
  atomic failure and preservation of earlier handles while the stop is live.
  Benchmark short and multi-megabyte strings with `-benchmem` at a fixed byte
  budget. Cover valid, escaped and late-malformed inputs. Time and allocated
  bytes should stay bounded as input length grows. Construct inputs outside
  the measured loop.
- **Wire.** Round-trip the exact valid key and propagate the producer's
  malformed-key error without a partial page. Force a delivery failure after
  hosted success, then retry during the same stop and check handle reuse.
- **ferretd and Editorium.** Repeat expansions with overlapping pages and
  exhaust DAP handles without evicting earlier ones. Exercise delayed replies
  across resume or close and verify that a new stop never displays the old
  children.
