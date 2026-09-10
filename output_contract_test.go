package api_test

import (
	"errors"
	"testing"

	"github.com/MontFerret/api"
)

func TestExecutionPreservesOutputPresence(t *testing.T) {
	failure := errors.New("execution or cleanup failed")
	populated := &api.Output{ContentType: "application/json", Content: []byte("42")}

	for _, tc := range []struct {
		name   string
		output *api.Output
		err    error
	}{
		{name: "absent output", err: failure},
		{name: "zero-valued output", output: &api.Output{}},
		{name: "zero-valued output with error", output: &api.Output{}, err: failure},
		{name: "populated output", output: populated},
		{name: "populated output with error", output: populated, err: failure},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var runtime api.Runtime = &outputRuntime{output: tc.output, err: tc.err}
			var session api.Session = &outputSession{output: tc.output, err: tc.err}

			t.Run("runtime", func(t *testing.T) {
				output, err := runtime.Run(t.Context(), api.NewAnonymousSource("RETURN 42"))
				if output != tc.output || err != tc.err {
					t.Fatalf("got output=%p err=%v, want output=%p err=%v", output, err, tc.output, tc.err)
				}
			})

			t.Run("session", func(t *testing.T) {
				output, err := session.Run(t.Context())
				if output != tc.output || err != tc.err {
					t.Fatalf("got output=%p err=%v, want output=%p err=%v", output, err, tc.output, tc.err)
				}
			})
		})
	}
}
