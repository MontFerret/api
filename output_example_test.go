package api_test

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"

	"github.com/MontFerret/api"
)

// collectQuery illustrates obtaining a handle, establishing fallback closure,
// and returning available content independently of the terminal error.
func collectQuery(ctx context.Context, runtime api.Runtime, src api.Source) (content *api.Content, err error) {
	output, err := runtime.Run(ctx, src)
	if err != nil {
		return nil, err
	}
	defer func() { err = errors.Join(err, output.Close()) }()
	return output.Collect(ctx)
}

// streamQuery illustrates forwarding borrowed chunks without retaining them.
// The writer must complete each write before returning from the callback.
func streamQuery(ctx context.Context, runtime api.Runtime, src api.Source, dst io.Writer) (err error) {
	output, err := runtime.Run(ctx, src)
	if err != nil {
		return err
	}
	defer func() { err = errors.Join(err, output.Close()) }()
	return output.Consume(ctx, func(ctx context.Context, chunk []byte) error {
		if err := ctx.Err(); err != nil {
			return err
		}
		n, err := dst.Write(chunk)
		if err != nil {
			return err
		}
		if n != len(chunk) {
			return io.ErrShortWrite
		}
		return nil
	})
}

func ExampleOutput_Collect() {
	// An application supplies its native or remote Runtime. This scripted runtime
	// illustrates caller handling of available content plus a terminal error.
	failure := errors.New("execution failed after a prefix")
	runtime := &outputRuntime{output: &scriptedOutput{
		content: &api.Content{
			Metadata: api.Metadata{ContentType: "application/json", Length: 2, LengthKnown: true},
			Data:     []byte("4"),
		},
		terminalErr: failure,
	}}
	content, err := collectQuery(context.Background(), runtime, api.NewAnonymousSource("RETURN 42"))
	if content != nil {
		fmt.Printf("%s: %q\n", content.Metadata.ContentType, content.Data)
	}
	if err != nil {
		fmt.Println("terminal failure:", errors.Is(err, failure))
	}
	// Output:
	// application/json: "4"
	// terminal failure: true
}

func ExampleOutput_Consume() {
	// Chunk boundaries are arbitrary. The example destination retains bytes so
	// its result can be shown; streamQuery itself does not accumulate the payload.
	runtime := &outputRuntime{output: &scriptedOutput{chunks: [][]byte{[]byte("4"), []byte("2")}}}
	var dst bytes.Buffer
	if err := streamQuery(context.Background(), runtime, api.NewAnonymousSource("RETURN 42"), &dst); err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(dst.String())
	// Output: 42
}
