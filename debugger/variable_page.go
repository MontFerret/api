package debugger

import (
	"errors"
	"math"
)

// VariableFilter selects the immediate children in a page.
type VariableFilter string

const (
	// VariableFilterAll selects named children first, followed by indexed children.
	VariableFilterAll VariableFilter = ""
	// VariableFilterNamed selects named children in stable order for the stop.
	VariableFilterNamed VariableFilter = "named"
	// VariableFilterIndexed selects indexed children in ascending source order.
	VariableFilterIndexed VariableFilter = "indexed"
)

const (
	// DefaultVariablePageSize is the number requested when Count is zero.
	DefaultVariablePageSize = 100
	// MaxVariablePageSize is the largest accepted Count for one call.
	MaxVariablePageSize = 1000
)

// ErrInvalidVariablePage identifies an unknown filter or an invalid page range.
var ErrInvalidVariablePage = errors.New("invalid debugger variable page")

// VariablePage requests a range in the selected child sequence. Start is a
// zero-based offset. Count is the maximum number of children to read. The
// zero value requests the first bounded page across both child kinds.
type VariablePage struct {
	Filter VariableFilter `json:"filter,omitempty"`
	Start  int64          `json:"start,omitempty"`
	Count  int64          `json:"count,omitempty"`
}

// Normalize replaces a zero Count with DefaultVariablePageSize. An unknown
// filter, negative Start or Count, Count above MaxVariablePageSize, or an
// exclusive end exceeding math.MaxInt64 returns ErrInvalidVariablePage.
func (p VariablePage) Normalize() (VariablePage, error) {
	switch p.Filter {
	case VariableFilterAll, VariableFilterNamed, VariableFilterIndexed:
	default:
		return VariablePage{}, ErrInvalidVariablePage
	}

	if p.Count == 0 {
		p.Count = DefaultVariablePageSize
	}

	if p.Start < 0 || p.Count < 0 || p.Count > MaxVariablePageSize || p.Start > math.MaxInt64-p.Count {
		return VariablePage{}, ErrInvalidVariablePage
	}

	return p, nil
}
