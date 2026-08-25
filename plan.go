package api

import (
	"context"
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
)
