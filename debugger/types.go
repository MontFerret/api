package debugger

import (
	"github.com/MontFerret/api/result"
	"github.com/MontFerret/api/source"
)

type (
	// PointID identifies a debug point within one compiled program.
	PointID int

	// BreakpointID identifies a breakpoint within one debugger session.
	BreakpointID int

	// FunctionID identifies a callable function within one compiled program.
	FunctionID int

	// BreakpointBindingMode selects how a requested source location resolves to
	// an executable debug point.
	BreakpointBindingMode int

	// Value describes a value without reading its children. Producing Type and
	// Display takes bounded work and space. Presentation and counts are
	// detached snapshots that remain readable after the stop.
	Value struct {
		// Type is the user-facing semantic type, including for host values.
		Type string `json:"type"`
		// Display is a cheap, bounded summary. Collection summaries can use a
		// cheaply known size without previewing elements.
		Display string `json:"display"`
		// Reference is positive when at least one paging mode is safe and bounded.
		// A named or indexed view may be available even if VariableFilterAll
		// is unsupported.
		// A known empty collection or a summary-only value has reference zero.
		Reference ValueReference `json:"reference"`

		// IndexedCount and NamedCount are optional, read-only exact counts of
		// each child kind. Nil means unknown. A present count is nonnegative
		// and can be zero. A known count alone does not grant expansion.
		IndexedCount *int64 `json:"indexedCount,omitempty"`
		NamedCount   *int64 `json:"namedCount,omitempty"`
	}

	// Variable describes a local, bind parameter, or immediate child of a
	// debugger value. A named child's Name preserves its valid UTF-8 key
	// exactly. An indexed child's Name is its absolute decimal index.
	Variable struct {
		Name    string `json:"name"`
		Value   Value  `json:"value"`
		Mutable bool   `json:"mutable"`
		Param   bool   `json:"param"`
	}

	// Frame describes the paused top frame or one of its callers.
	Frame struct {
		Name       string          `json:"name"`
		Location   source.Location `json:"location"`
		FunctionID FunctionID      `json:"functionID"`
	}

	// Breakpoint describes a requested source-location breakpoint and its resolved
	// executable location, when one exists.
	Breakpoint struct {
		Location          source.Range          `json:"location"`
		RequestedLocation source.Location       `json:"requestedLocation"`
		ID                BreakpointID          `json:"id"`
		PointID           PointID               `json:"pointID"`
		FunctionID        FunctionID            `json:"functionID"`
		BindingMode       BreakpointBindingMode `json:"bindingMode"`
		Bound             bool                  `json:"bound"`
	}

	// BreakpointRequest specifies one position and its binding policy within a
	// source-wide replacement. Source identity belongs to ReplaceBreakpoints.
	BreakpointRequest struct {
		Position source.Position   `json:"position"`
		Options  BreakpointOptions `json:"options"`
	}

	// BreakpointOptions configures how a requested source location binds.
	BreakpointOptions struct {
		BindingMode BreakpointBindingMode `json:"bindingMode"`
	}

	// Event reports a debugger stop, completion, or termination.
	// Retained data is materialized; reading an event does not consume another
	// observer's output. Available completion content may accompany an error.
	Event struct {
		Error error `json:"error"`

		// Output is detached encoded content, not a live consumable handle.
		// Nil means absent; a non-nil pointer means present, including empty data.
		// Retained snapshots must remain valid after debugger cleanup and must
		// not share mutable implementation buffers. Recipients coordinate any
		// mutation of shared Content themselves.
		Output           *result.Content `json:"output"`
		Reason           Reason          `json:"reason"`
		HitBreakpointIDs []BreakpointID  `json:"hitBreakpointIDs"`
		Location         source.Range    `json:"location"`
		Depth            int             `json:"depth"`
	}
)

const (
	// BreakpointBindNextExecutableInSource selects the next executable point in
	// the named source and is the zero-value default.
	BreakpointBindNextExecutableInSource BreakpointBindingMode = iota
	BreakpointBindExact
	BreakpointBindNextExecutableInFunction
)

// NoFunction identifies the top-level program body rather than a callable function.
const NoFunction FunctionID = -1

func BreakpointBindingModeFromString(s string) BreakpointBindingMode {
	switch s {
	case "next-executable-in-source":
		return BreakpointBindNextExecutableInSource
	case "exact":
		return BreakpointBindExact
	case "next-executable-in-function":
		return BreakpointBindNextExecutableInFunction
	default:
		return BreakpointBindNextExecutableInSource
	}
}
