package result

import (
	"context"
	"io"
)

type (
	// Metadata describes the encoded representation of a complete payload.
	// It does not establish whether content is present. Output metadata is an
	// immutable descriptor, not a remaining-byte count or progress counter.
	Metadata struct {
		// ContentType identifies the encoded representation's media type.
		ContentType string `json:"contentType"`

		// Length is the exact total payload length in bytes, as exposed
		// to the consumer, not a record count or transport-frame size.
		// It is meaningful only when LengthKnown is true. Implementations
		// must not require unrestricted preallocation based on this value.
		Length int64 `json:"length"`

		// LengthKnown distinguishes unknown length from known zero length.
		// Unknown length is normal; it must not require buffering the payload
		// to discover its size. When true, Length must be nonnegative.
		// Known zero length does not establish content presence.
		LengthKnown bool `json:"lengthKnown"`
	}

	// Content is detached, materialized encoded data owned by its recipient.
	// It remains valid after the Output and its resources are closed. A nil
	// *Content means no content is available; a non-nil pointer, including
	// &Content{}, means content is present regardless of metadata or data length.
	//
	// JSON nests the descriptor under "metadata" and encodes Data under "data"
	// using the standard byte-slice representation, without interpreting the
	// payload as JSON. Nil and empty Data slices encode as null and "", respectively;
	// neither means the containing Content is absent.
	Content struct {
		// Metadata preserves the original complete-payload descriptor, even
		// when Data contains only a prefix received before failure.
		Metadata Metadata `json:"metadata"`

		// Data contains the bytes actually collected. On successful complete
		// collection, its length must match Metadata.Length when known.
		Data []byte `json:"data"`
	}

	// Consumer receives ordered encoded chunks synchronously and without overlap.
	// Chunks are borrowed, read-only, and valid only during the call; copy bytes
	// that must be retained. Boundaries need not align with values, lines, records,
	// or character boundaries. Empty chunks are not EOF.
	//
	// ctx is the effective consumption context, bounded by both the invocation
	// and consumption contexts. Returning an error stops further delivery and
	// initiates finalization. Cancellation cannot forcibly interrupt arbitrary
	// callback code or invalidate borrowed buffers while the callback uses them.
	Consumer func(ctx context.Context, chunk []byte) error

	// Output is a caller-owned, one-shot handle to encoded query output.
	// Consume and Collect are alternative terminal operations. A handle is usable
	// even when no content is available; content presence is determined during
	// consumption, not by the handle, metadata, or byte-slice length.
	//
	// Execution starts before the handle is returned; it is not deferred until
	// consumption. Terminal execution, encoding, delivery, and output-owned cleanup
	// errors are observed through Consume or Collect. Callers running queries only
	// for side effects must still consume to observe completion. Streaming-capable
	// consumption does not require incremental query evaluation or native encoding.
	// Serialize detached Content, not a live Output handle.
	//
	// Consumption admission validates a non-nil context, then a non-nil Consumer
	// for Consume, then cancellation of the consumption and invocation contexts,
	// before atomically claiming the handle. Invalid or already-canceled calls
	// return errors without claiming it or disturbing another consumer. A fresh
	// context can retry a rejected call only within the invocation's lifetime.
	// After admission, cancellation or any other failure finalizes the output;
	// consumption cannot be resumed or repeated.
	//
	// Valid competing consumption attempts match ErrInUse while another consumer
	// owns delivery and stopping has not begun. Once closure or finalization begins,
	// new attempts match ErrClosed. Rejected calls never disrupt active consumption.
	// Metadata may be read concurrently, including from a consumer. Reentrant
	// consumption is rejected by the same admission rules. A consumer must not call
	// Close synchronously on this output because Close waits for that callback.
	//
	// The invocation context bounds the output's lifetime. The consumption context
	// may shorten but never extend or revive it. The effective context observes both
	// cancellations and their earliest deadline. Cancellation errors preserve
	// context.Canceled and context.DeadlineExceeded through errors.Is.
	//
	// Admitted operations finalize owned resources before returning, including on
	// failure and during consumer panic unwinding, without swallowing the panic.
	// Errors preserve all observed execution, consumer, and cleanup causes through
	// standard error traversal. Cleanup errors remain available through Close.
	Output interface {
		// Close abandons unread output without draining arbitrarily large content
		// and releases owned resources. It does not certify successful execution.
		// During active consumption it requests cancellation and waits for the
		// active callback and finalization; no borrowed buffer is released early.
		//
		// If explicit closure wins before the consumption outcome is committed,
		// that operation's error matches ErrClosed and preserves other observed
		// failures. Closure after commitment only waits for cleanup. Automatic
		// finalization does not itself add ErrClosed to the original operation;
		// caller-context cancellation remains identifiable as a context error.
		//
		// Close is safe concurrently and idempotent, returning the recorded cleanup
		// outcome, usually nil, even after finalization. Repeated calls need not
		// return identical error-wrapper pointers. Being closed is not itself a
		// Close error. It does not close caller-owned sessions or borrowed parents.
		io.Closer

		// Metadata returns the immutable complete-payload descriptor locally,
		// without I/O, including during consumption and after closure. Run must
		// establish it before exposing the handle without waiting for or buffering
		// the complete payload solely to determine length.
		Metadata() Metadata

		// Consume delivers borrowed chunks according to the admission and lifetime
		// rules above, then finalizes resources before returning its terminal error.
		// No callback remains active or is invoked after return. Absent content
		// invokes no callback; present empty content invokes at least one empty
		// chunk. A consumer error stops further delivery and is preserved alongside
		// other observed failures. An otherwise complete delivery whose total bytes
		// differ from a known length must return an error.
		Consume(ctx context.Context, consumer Consumer) error

		// Collect materializes detached content under the same admission, lifetime,
		// and finalization rules as Consume. Nil content means none is available;
		// present empty content returns a non-nil pointer. Available content,
		// including a received prefix, is returned alongside any terminal error.
		// Metadata remains the original descriptor; len(Data) is the collected size.
		// Complete successful collection must match a known length.
		//
		// Implementations may transfer suitable owned bytes directly rather than
		// copying them through a generic chunk accumulator. Detached data must not
		// alias borrowed or reusable implementation buffers.
		Collect(ctx context.Context) (*Content, error)
	}
)
