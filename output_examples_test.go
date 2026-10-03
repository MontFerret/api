package api_test

import (
	"bytes"
	"errors"
	"io"
	"testing"

	"github.com/MontFerret/api"
)

type writerFunc func([]byte) (int, error)

func (f writerFunc) Write(data []byte) (int, error) { return f(data) }

// These tests exercise the example's writer and error handling, not the output
// adapter's lifecycle. Scripted delivery is intentional; adapter tests belong
// in the native and remote implementation repositories.
func TestStreamQueryHandlesWriterAndTerminalFailures(t *testing.T) {
	destinationFailure := errors.New("destination failure")
	executionFailure := errors.New("terminal execution failure")
	cleanupFailure := errors.New("output cleanup failure")
	for _, tc := range []struct {
		name        string
		writer      io.Writer
		terminalErr error
		cleanupErr  error
		causes      []error
		wantWrites  int
	}{
		{name: "complete writes", writer: io.Discard, wantWrites: 2},
		{name: "short write", writer: writerFunc(func(data []byte) (int, error) { return len(data) - 1, nil }), causes: []error{io.ErrShortWrite}, wantWrites: 1},
		{name: "writer failure", writer: writerFunc(func([]byte) (int, error) { return 0, destinationFailure }), causes: []error{destinationFailure}, wantWrites: 1},
		{name: "partial write with cause", writer: writerFunc(func(data []byte) (int, error) { return len(data) - 1, destinationFailure }), causes: []error{destinationFailure}, wantWrites: 1},
		{name: "terminal failure", writer: io.Discard, terminalErr: executionFailure, causes: []error{executionFailure}, wantWrites: 2},
		{name: "cleanup failure", writer: io.Discard, cleanupErr: cleanupFailure, causes: []error{cleanupFailure}, wantWrites: 2},
		{name: "multiple failures", writer: writerFunc(func([]byte) (int, error) { return 0, destinationFailure }), terminalErr: executionFailure, cleanupErr: cleanupFailure, causes: []error{destinationFailure, executionFailure, cleanupFailure}, wantWrites: 1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			output := &scriptedOutput{chunks: [][]byte{[]byte("first"), []byte("second")}, terminalErr: tc.terminalErr, cleanupErr: tc.cleanupErr}
			writes := 0
			dst := writerFunc(func(data []byte) (int, error) {
				writes++
				return tc.writer.Write(data)
			})
			err := streamQuery(t.Context(), &outputRuntime{output: output}, api.NewAnonymousSource("RETURN 42"), dst)
			if len(tc.causes) == 0 && err != nil {
				t.Fatalf("unexpected forwarding error: %v", err)
			}
			for _, cause := range tc.causes {
				if !errors.Is(err, cause) {
					t.Fatalf("forwarding error %v lost cause %v", err, cause)
				}
			}
			if writes != tc.wantWrites || output.consumeCalls != 1 || output.collectCalls != 0 || output.closeCalls != 1 {
				t.Fatalf("caller calls: writes=%d consume=%d collect=%d fallback close=%d", writes, output.consumeCalls, output.collectCalls, output.closeCalls)
			}
		})
	}
}

func TestStreamQueryReconstructsChunks(t *testing.T) {
	var dst bytes.Buffer
	output := &scriptedOutput{chunks: [][]byte{{0, 0xff}, {}, {'x'}, {0x80, '\n'}}}
	err := streamQuery(t.Context(), &outputRuntime{output: output}, api.NewAnonymousSource("RETURN 42"), &dst)
	if err != nil {
		t.Fatal(err)
	}
	if want := []byte{0, 0xff, 'x', 0x80, '\n'}; !bytes.Equal(dst.Bytes(), want) {
		t.Fatalf("forwarded bytes = %v, want %v", dst.Bytes(), want)
	}
}

func TestQueryExamplesReturnPreparationErrors(t *testing.T) {
	failure := errors.New("preparation failure")
	runtime := &outputRuntime{err: failure}
	content, err := collectQuery(t.Context(), runtime, api.NewAnonymousSource("RETURN 42"))
	if content != nil || !errors.Is(err, failure) {
		t.Fatalf("collection preparation result = (%v, %v)", content, err)
	}
	if err := streamQuery(t.Context(), runtime, api.NewAnonymousSource("RETURN 42"), io.Discard); !errors.Is(err, failure) {
		t.Fatalf("streaming preparation error = %v", err)
	}
}
