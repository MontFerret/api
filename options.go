package api

import "fmt"

type (
	PlanOptions interface {
		SetOptimizationLevel(OptimizationLevel) error
	}

	PlanOption = func(PlanOptions) error

	// SessionOptions configures a session directly or queues implementation options.
	// Setters may report portable or application-level errors immediately and defer
	// runtime-specific conversion and validation until the setting is used. A nil
	// setter error does not certify that the runtime can use the value.
	//
	// Invalid settings must fail the operation no later than their relevant point of use.
	// Validation need not precede compilation, resource acquisition, or all query
	// execution. Output codec availability may be checked during result encoding,
	// after the query has run. Implementations document validation and mutable-input
	// conversion or snapshot timing.
	//
	// Later setters override earlier values; SetParams merges keys.
	// Runtime-specific options reject incompatible targets.
	SessionOptions interface {
		SetParam(string, any) error
		SetParams(map[string]any) error
		SetOutputContentType(string) error
		SetFSRoot(string) error
	}

	// SessionOption applies portable or application-specific session configuration.
	// Implementations invoke non-nil callbacks once in order and join their returned
	// errors. Callbacks may return errors immediately; runtime-specific validation
	// may instead fail in the operation that uses the setting. See SessionOptions.
	SessionOption = func(SessionOptions) error
)

// WithOptimizationLevel sets the optimization level for the execution plan.
// The callback rejects unknown enum values before invoking the target setter;
// the runtime determines which known levels it supports.
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
// The runtime owns conversion and validation of the host value.
func WithParam(key string, value any) SessionOption {
	return func(opts SessionOptions) error {
		return opts.SetParam(key, value)
	}
}

// WithParams merges the provided parameter map into the session environment,
// overriding existing keys while preserving any other previously defined parameters.
// The runtime defines when values are converted or snapshotted.
func WithParams(params map[string]any) SessionOption {
	return func(opts SessionOptions) error {
		return opts.SetParams(params)
	}
}

// WithOutputContentType selects the output codec content type for session results.
// Codec availability may be checked when output is encoded, after query execution.
func WithOutputContentType(contentType string) SessionOption {
	return func(opts SessionOptions) error {
		return opts.SetOutputContentType(contentType)
	}
}

// WithFSRoot selects the rooted filesystem used by one execution session.
// The runtime defines path validation and owns any filesystem resources created
// for the session. Validation and construction may occur when the root is applied.
func WithFSRoot(root string) SessionOption {
	return func(opts SessionOptions) error {
		return opts.SetFSRoot(root)
	}
}
