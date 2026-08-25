package api

import (
	"context"
	"io"

	"github.com/MontFerret/api/debugger"
	"github.com/MontFerret/api/result"
	"github.com/MontFerret/api/source"
)

type (
	Runtime interface {
		io.Closer
		Run(ctx context.Context, src source.File, opts ...SessionOption) (result.Output, error)
		Compile(ctx context.Context, src source.File) (Plan, error)
		CompileDebug(ctx context.Context, src source.File) (Plan, error)
	}

	Plan interface {
		io.Closer
		Params() []string
		NewSession(ctx context.Context, opts ...SessionOption) (Session, error)
		NewDebugSession(ctx context.Context, opts ...SessionOption) (debugger.Session, error)
	}

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
