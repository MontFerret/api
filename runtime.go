package api

import (
	"context"
	"io"

	"github.com/MontFerret/api/result"
	"github.com/MontFerret/api/source"
)

type Runtime interface {
	io.Closer
	Run(ctx context.Context, src source.File, opts ...SessionOption) (result.Output, error)
	Compile(ctx context.Context, src source.File, opts ...PlanOption) (Plan, error)
	CompileDebug(ctx context.Context, src source.File, opts ...PlanOption) (Plan, error)
}
