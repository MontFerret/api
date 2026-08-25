package api

import (
	"context"
	"fmt"
	"io"

	"github.com/MontFerret/api/result"
)

type (
	Session interface {
		io.Closer
		Run(c context.Context) (result.Output, error)
	}

	SessionOptions interface {
		SetParam(string, any) error
		SetOutputContentType(string) error
	}

	SessionOption func(SessionOptions) error
)

// WithSessionParam sets a session parameter for the execution.
func WithSessionParam(key string, value any) SessionOption {
	return func(opts SessionOptions) error {
		return opts.SetParam(key, value)
	}
}

// WithSessionParams merges the provided parameter map into the session environment,
// overriding existing keys while preserving any other previously defined parameters.
func WithSessionParams(params map[string]any) SessionOption {
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
