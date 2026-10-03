package api

import (
	"github.com/MontFerret/api/result"
	"github.com/MontFerret/api/source"
)

type (
	// Source represents the input data for a Ferret query.
	Source = source.Source

	// Position represents a specific point in a source file, defined by line and column numbers.
	Position = source.Position

	// Span represents a range of characters in a source file, defined by start and end positions.
	Span = source.Span

	// Location represents the location of a specific point in a source file, including the file name and position.
	Location = source.Location

	// Range represents a range of characters in a source file, including the location and span.
	Range = source.Range

	// Metadata describes the complete encoded payload independently of content
	// presence. See result.Metadata for length and immutability requirements.
	Metadata = result.Metadata

	// Content is detached, caller-owned encoded data returned by Output.Collect
	// or retained debugger events. A nil *Content means no content is available;
	// a non-nil pointer, including &Content{}, means present content. See result.Content.
	Content = result.Content

	// Consumer receives borrowed, read-only chunks during Output.Consume.
	// Copy retained bytes; see result.Consumer for callback and context requirements.
	Consumer = result.Consumer

	// Output is the caller-owned, one-shot consumable handle returned by execution.
	// Consume or Collect observes completion and finalizes resources; Close abandons
	// unread output. See result.Output for lifetime, presence, and error requirements.
	Output = result.Output
)

var (
	// ErrOutputInUse is result.ErrInUse. Use errors.Is to identify a consumption
	// attempt rejected because another consumer owns the output.
	ErrOutputInUse = result.ErrInUse

	// ErrOutputClosed is result.ErrClosed. Use errors.Is to identify consumption
	// rejected after stopping begins, or interrupted by an explicit Close that
	// wins before outcome commitment. Close itself returns its cleanup outcome.
	ErrOutputClosed = result.ErrClosed
)

// NewSource creates a new Source instance with the given name and content.
func NewSource(name, content string) Source {
	return source.New(name, content)
}

// NewAnonymousSource creates a new anonymous Source instance with the given content.
func NewAnonymousSource(content string) Source {
	return source.NewAnonymous(content)
}
