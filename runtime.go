package api

import (
	"context"
	"io"

	"github.com/MontFerret/api/result"
)

type Runtime interface {
	io.Closer
	Run(ctx context.Context, src Source, opts ...SessionOption) (result.Output, error)
	Compile(ctx context.Context, src Source, opts ...PlanOption) (Plan, error)
	CompileDebug(ctx context.Context, src Source, opts ...PlanOption) (Plan, error)
}
