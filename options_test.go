package api

import (
	"errors"
	"testing"
)

func TestWithFSRootForwardsToSessionOptions(t *testing.T) {
	t.Parallel()

	target := &sessionOptionsSpy{}
	if err := WithFSRoot("/runtime")(target); err != nil {
		t.Fatalf("WithFSRoot() error = %v", err)
	}

	if target.fsRoot != "/runtime" {
		t.Fatalf("filesystem root = %q, want /runtime", target.fsRoot)
	}
}

func TestWithFSRootPreservesTargetError(t *testing.T) {
	t.Parallel()

	want := errors.New("invalid filesystem root")
	target := &sessionOptionsSpy{fsRootErr: want}
	err := WithFSRoot("/runtime")(target)
	if !errors.Is(err, want) {
		t.Fatalf("WithFSRoot() error = %v, want %v", err, want)
	}
}

type sessionOptionsSpy struct {
	fsRoot    string
	fsRootErr error
}

func (*sessionOptionsSpy) SetParam(string, any) error {
	return nil
}

func (*sessionOptionsSpy) SetParams(map[string]any) error {
	return nil
}

func (*sessionOptionsSpy) SetOutputContentType(string) error {
	return nil
}

func (s *sessionOptionsSpy) SetFSRoot(root string) error {
	s.fsRoot = root

	return s.fsRootErr
}
