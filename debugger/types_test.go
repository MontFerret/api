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

func (sessionContract) Pause() error { return nil }

func (sessionContract) SetBreakpoint(source.Location) (debugger.Breakpoint, error) {
	return debugger.Breakpoint{}, nil
}

func (sessionContract) SetBreakpointAt(source.Location, debugger.BreakpointOptions) (debugger.Breakpoint, error) {
	return debugger.Breakpoint{}, nil
}

func (sessionContract) DeleteBreakpoint(debugger.BreakpointID) error { return nil }

func (sessionContract) Breakpoints() []debugger.Breakpoint { return nil }

func (sessionContract) Frames() ([]debugger.Frame, error) { return nil, nil }

func (sessionContract) Locals() ([]debugger.Variable, error) { return nil, nil }

func (sessionContract) FrameLocals(int) ([]debugger.Variable, error) { return nil, nil }

func (sessionContract) Variables(debugger.ValueReference) ([]debugger.Variable, error) {
	return nil, nil
}

func (sessionContract) Evaluate(context.Context, string) (debugger.Value, error) {
	return debugger.Value{}, nil
}

func (sessionContract) EvaluateFrame(context.Context, int, string) (debugger.Value, error) {
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
