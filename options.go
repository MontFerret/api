package api

import "fmt"

type (
	PlanOptions interface {
		SetOptimizationLevel(OptimizationLevel) error
	}

	PlanOption = func(PlanOptions) error

	SessionOptions interface {
		SetParam(string, any) error
		SetParams(map[string]any) error
		SetOutputContentType(string) error
	}

	SessionOption = func(SessionOptions) error
)

// WithOptimizationLevel sets the optimization level for the execution plan.
func WithOptimizationLevel(level OptimizationLevel) PlanOption {
	return func(opts PlanOptions) error {
		switch level {
		case OptimizationNone,
			OptimizationBasic,
			OptimizationFull,
			OptimizationAggressive:
			return opts.SetOptimizationLevel(level)
		default:
			return fmt.Errorf("invalid optimization level: %d", level)
		}
	}
}

// WithParam sets a session parameter for the execution.
func WithParam(key string, value any) SessionOption {
	return func(opts SessionOptions) error {
		return opts.SetParam(key, value)
	}
}

// WithParams merges the provided parameter map into the session environment,
// overriding existing keys while preserving any other previously defined parameters.
func WithParams(params map[string]any) SessionOption {
	return func(opts SessionOptions) error {
		return opts.SetParams(params)
	}
}

// WithOutputContentType selects the output codec content type for session results.
func WithOutputContentType(contentType string) SessionOption {
	return func(opts SessionOptions) error {
		return opts.SetOutputContentType(contentType)
	}
}
