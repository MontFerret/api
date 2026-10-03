package api_test

import (
	"context"
	"errors"

	"github.com/MontFerret/api"
)

type (
	outputRuntime struct {
		output api.Output
		err    error
	}

	outputSession struct {
		output api.Output
		err    error
	}

	// scriptedOutput supplies predetermined outcomes to caller examples. It is
	// deliberately not a state machine or an adapter conformance implementation:
	// these tests verify caller handling, not one-shot, lifetime, or cleanup rules.
	scriptedOutput struct {
		content      *api.Content
		chunks       [][]byte
		terminalErr  error
		cleanupErr   error
		collectCalls int
		consumeCalls int
		closeCalls   int
	}
)

var (
	_ api.Runtime = (*outputRuntime)(nil)
	_ api.Session = (*outputSession)(nil)
	_ api.Output  = (*scriptedOutput)(nil)
)

func (r *outputRuntime) Version(context.Context) (api.Version, error) {
	return api.Version(""), nil
}

func (r *outputRuntime) Run(context.Context, api.Source, ...api.SessionOption) (api.Output, error) {
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

func (s *outputSession) Run(context.Context) (api.Output, error) {
	return s.output, s.err
}

func (s *outputSession) Close() error {
	return nil
}

func (o *scriptedOutput) Metadata() api.Metadata {
	if o.content == nil {
		return api.Metadata{}
	}
	return o.content.Metadata
}

func (o *scriptedOutput) Consume(ctx context.Context, consumer api.Consumer) error {
	o.consumeCalls++
	chunks := o.chunks
	if chunks == nil && o.content != nil {
		chunks = [][]byte{o.content.Data}
	}
	for _, chunk := range chunks {
		if err := consumer(ctx, chunk); err != nil {
			return errors.Join(err, o.terminalErr, o.cleanupErr)
		}
	}
	return errors.Join(o.terminalErr, o.cleanupErr)
}

func (o *scriptedOutput) Collect(context.Context) (*api.Content, error) {
	o.collectCalls++
	return o.content, errors.Join(o.terminalErr, o.cleanupErr)
}

func (o *scriptedOutput) Close() error {
	o.closeCalls++
	return o.cleanupErr
}
