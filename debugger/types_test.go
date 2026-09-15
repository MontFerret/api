package debugger_test

import (
	"context"
	"testing"

	"github.com/MontFerret/api/debugger"
	"github.com/MontFerret/api/source"
)

type sessionContract struct{}

func (sessionContract) Start(context.Context) (*debugger.Event, error) { return nil, nil }

func (sessionContract) Continue(context.Context) (*debugger.Event, error) { return nil, nil }

func (sessionContract) StepIn(context.Context) (*debugger.Event, error) { return nil, nil }

func (sessionContract) StepOver(context.Context) (*debugger.Event, error) { return nil, nil }

func (sessionContract) StepOut(context.Context) (*debugger.Event, error) { return nil, nil }

func (sessionContract) Pause(ctx context.Context) error { return nil }

func (sessionContract) ReplaceBreakpoints(context.Context, string, []debugger.BreakpointRequest) ([]debugger.Breakpoint, error) {
	return nil, nil
}

func (sessionContract) SetBreakpoint(ctx context.Context, loc source.Location) (debugger.Breakpoint, error) {
	return debugger.Breakpoint{}, nil
}

func (sessionContract) SetBreakpointAt(ctx context.Context, loc source.Location, opts debugger.BreakpointOptions) (debugger.Breakpoint, error) {
	return debugger.Breakpoint{}, nil
}

func (sessionContract) DeleteBreakpoint(ctx context.Context, id debugger.BreakpointID) error {
	return nil
}

func (sessionContract) Breakpoints(context.Context) ([]debugger.Breakpoint, error) { return nil, nil }

func (sessionContract) Frames(ctx context.Context) ([]debugger.Frame, error) { return nil, nil }

func (sessionContract) Locals(ctx context.Context) ([]debugger.Variable, error) { return nil, nil }

func (sessionContract) FrameLocals(ctx context.Context, frame int) ([]debugger.Variable, error) {
	return nil, nil
}

func (sessionContract) Variables(ctx context.Context, reference debugger.ValueReference) ([]debugger.Variable, error) {
	return nil, nil
}

func (sessionContract) Evaluate(ctx context.Context, expression string) (debugger.Value, error) {
	return debugger.Value{}, nil
}

func (sessionContract) EvaluateFrame(ctx context.Context, frame int, expression string) (debugger.Value, error) {
	return debugger.Value{}, nil
}

func (sessionContract) Close() error { return nil }

func TestDebuggerSessionUsesConventionalStepNames(t *testing.T) {
	var session debugger.Session = sessionContract{}

	if _, err := session.StepIn(context.Background()); err != nil {
		t.Fatal(err)
	}
	if _, err := session.StepOver(context.Background()); err != nil {
		t.Fatal(err)
	}
	if _, err := session.StepOut(context.Background()); err != nil {
		t.Fatal(err)
	}
}

func TestBreakpointBindingModeUsesSourceTerminology(t *testing.T) {
	if got := debugger.BreakpointBindingModeFromString("next-executable-in-source"); got != debugger.BreakpointBindNextExecutableInSource {
		t.Fatalf("binding mode = %v", got)
	}
	if got := debugger.BreakpointBindingModeFromString("unknown"); got != debugger.BreakpointBindNextExecutableInSource {
		t.Fatalf("default binding mode = %v", got)
	}
}

func TestPortableIdentityConventions(t *testing.T) {
	if debugger.NoFunction != -1 {
		t.Fatal("unexpected top-level identity")
	}
	for _, value := range []debugger.ValueReference{-1, 0, 1, 42} {
		if value.Valid() != (value > 0) {
			t.Fatalf("reference %d validity", value)
		}
	}
}

func TestDebuggerSessionRequiresSourceReplacement(t *testing.T) {
	var session debugger.Session = sessionContract{}
	requests := []debugger.BreakpointRequest{{
		Position: source.Position{Line: 1},
		Options:  debugger.BreakpointOptions{BindingMode: debugger.BreakpointBindExact},
	}}
	if _, err := session.ReplaceBreakpoints(t.Context(), "buffer://query", requests); err != nil {
		t.Fatal(err)
	}
}
