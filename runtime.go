package api

import (
	"context"
	"io"
)

type Runtime interface {
	io.Closer
	Run(ctx context.Context, src Source, opts ...SessionOption) (Output, error)
	Compile(ctx context.Context, src Source, opts ...PlanOption) (Plan, error)
	CompileDebug(ctx context.Context, src Source, opts ...PlanOption) (Plan, error)
}
