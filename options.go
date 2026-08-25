package api

import "fmt"

type (
	PlanOptions interface {
		SetOptimizationLevel(OptimizationLevel) error
	}

	PlanOption func(PlanOptions) error

	SessionOptions interface {
		SetParam(string, any) error
		SetOutputContentType(string) error
	}

	SessionOption func(SessionOptions) error
)

// WithOptimizationLevel sets the optimization level for the execution plan.
func WithOptimizationLevel(level OptimizationLevel) PlanOption {
	return func(opts PlanOptions) error {
		if _, err := ParseOptimizationLevel(int(level)); err != nil {
			return err
		}

		if err := opts.SetOptimizationLevel(level); err != nil {
			return fmt.Errorf("failed to set optimization level: %w", err)
		}

		return nil
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
		for k, v := range params {
			if err := opts.SetParam(k, v); err != nil {
				return fmt.Errorf("failed to set session param %q: %w", k, err)
			}
		}

		return nil
	}
}

// WithOutputContentType selects the output codec content type for session results.
func WithOutputContentType(contentType string) SessionOption {
	return func(opts SessionOptions) error {
		if err := opts.SetOutputContentType(contentType); err != nil {
			return fmt.Errorf("failed to set output content type: %w", err)
		}

		return nil
	}
}
