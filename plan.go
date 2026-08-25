package api

import (
	"context"
	"fmt"
	"io"

	"github.com/MontFerret/api/debugger"
)

type (
	Plan interface {
		io.Closer
		Params() []string
		NewSession(ctx context.Context, opts ...SessionOption) (Session, error)
		NewDebugSession(ctx context.Context, opts ...SessionOption) (debugger.Session, error)
	}

	OptimizationLevel int

	PlanOptions interface {
		SetOptimizationLevel(OptimizationLevel) error
	}

	PlanOption func(PlanOptions) error
)

const (
	OptimizationNone OptimizationLevel = iota
	OptimizationBasic
	OptimizationFull
	OptimizationAggressive
)

func OptimizationLevelFromInt(level int) OptimizationLevel {
	switch level {
	case 0:
		return OptimizationNone
	case 1:
		return OptimizationBasic
	case 2:
		return OptimizationFull
	case 3:
		return OptimizationAggressive
	default:
		return OptimizationNone
	}
}

func WithOptimizationLevel(level OptimizationLevel) PlanOption {
	return func(opts PlanOptions) error {
		if err := opts.SetOptimizationLevel(level); err != nil {
			return fmt.Errorf("failed to set optimization level: %w", err)
		}

		return nil
	}
}
