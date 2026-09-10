package api

type (
	planOptionsSpy struct {
		err   error
		level OptimizationLevel
		calls int
	}

	sessionOptionsSpy struct {
		err    error
		method string
		args   []any
		calls  int
	}
)

func (s *planOptionsSpy) SetOptimizationLevel(level OptimizationLevel) error {
	s.level = level
	s.calls++

	return s.err
}

func (s *sessionOptionsSpy) SetParam(name string, value any) error {
	return s.record("param", name, value)
}

func (s *sessionOptionsSpy) SetParams(params map[string]any) error {
	return s.record("params", params)
}

func (s *sessionOptionsSpy) SetOutputContentType(contentType string) error {
	return s.record("content type", contentType)
}

func (s *sessionOptionsSpy) SetFSRoot(root string) error {
	return s.record("fs root", root)
}

func (s *sessionOptionsSpy) record(method string, args ...any) error {
	s.method = method
	s.args = args
	s.calls++

	return s.err
}
