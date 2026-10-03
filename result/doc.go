// Package result defines consumable encoded Output handles and detached Content.
//
// Output is live and one-shot: Consume receives borrowed chunks, Collect obtains
// detached materialized bytes, and Close abandons unread output. Metadata describes
// the complete encoded representation locally and remains unchanged after closure.
// A usable handle does not imply content presence or completed execution; terminal
// errors and any available content are observed during consumption.
//
// This package declares contracts and data, not an execution or transport adapter.
// Consumable delivery permits bounded buffers without requiring incremental query
// evaluation or encoding. Only detached Content and Metadata are serialized.
package result
