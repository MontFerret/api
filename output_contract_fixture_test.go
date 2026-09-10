package api_test

import (
	"context"

	"github.com/MontFerret/api"
)

type (
	outputRuntime struct {
		output *api.Output
		err    error
	}

	outputSession struct {
		output *api.Output
		err    error
	}
)

var (
	_ api.Runtime = (*outputRuntime)(nil)
	_ api.Session = (*outputSession)(nil)
)

func (r *outputRuntime) Run(context.Context, api.Source, ...api.SessionOption) (*api.Output, error) {
	return r.output, r.err
}

func (r *outputRuntime) Compile(context.Context, api.Source, ...api.PlanOption) (api.Plan, error) {
	return nil, nil
}

func (r *outputRuntime) CompileDebug(context.Context, api.Source, ...api.PlanOption) (api.Plan, error) {
	return nil, nil
}

func (r *outputRuntime) Close() error {
	return nil
}

func (s *outputSession) Run(context.Context) (*api.Output, error) {
	return s.output, s.err
}

func (s *outputSession) Close() error {
	return nil
}
