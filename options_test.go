package api

import (
	"errors"
	"fmt"
	"reflect"
	"testing"
)

func TestWithOptimizationLevelRejectsUnknownEnumBeforeSetter(t *testing.T) {
	t.Parallel()

	for _, level := range []OptimizationLevel{-1, OptimizationAggressive + 1} {
		t.Run(fmt.Sprint(level), func(t *testing.T) {
			target := &planOptionsSpy{}
			if err := WithOptimizationLevel(level)(target); err == nil {
				t.Fatal("unknown optimization level accepted")
			}

			if target.calls != 0 {
				t.Fatalf("setter calls = %d, want 0", target.calls)
			}
		})
	}
}

func TestWithOptimizationLevelForwardsKnownEnumsAndTargetErrors(t *testing.T) {
	t.Parallel()

	failure := errors.New("runtime does not support this optimization")
	for _, level := range []OptimizationLevel{OptimizationNone, OptimizationBasic, OptimizationFull, OptimizationAggressive} {
		t.Run(fmt.Sprint(level), func(t *testing.T) {
			for _, wantErr := range []error{nil, failure} {
				target := &planOptionsSpy{err: wantErr}
				if err := WithOptimizationLevel(level)(target); err != wantErr {
					t.Fatalf("error = %v, want unchanged target error %v", err, wantErr)
				}

				if target.calls != 1 || target.level != level {
					t.Fatalf("setter calls = %d, level = %d; want 1, %d", target.calls, target.level, level)
				}
			}
		})
	}
}

func TestSessionOptionHelpersForwardValuesAndTargetErrors(t *testing.T) {
	t.Parallel()

	value := make(chan int)
	params := map[string]any{"value": value}
	failure := errors.New("runtime option failure")
	for _, tc := range []struct {
		name   string
		option SessionOption
		method string
		args   []any
	}{
		{"parameter", WithParam("", value), "param", []any{"", value}},
		{"parameters", WithParams(params), "params", []any{params}},
		{"output codec", WithOutputContentType("unknown/codec"), "content type", []any{"unknown/codec"}},
		{"blank content type", WithOutputContentType(" \t"), "content type", []any{" \t"}},
		{"filesystem root", WithFSRoot("/runtime"), "fs root", []any{"/runtime"}},
		{"blank filesystem root", WithFSRoot(" \t"), "fs root", []any{" \t"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			for _, wantErr := range []error{nil, failure} {
				target := &sessionOptionsSpy{err: wantErr}
				if err := tc.option(target); err != wantErr {
					t.Fatalf("error = %v, want unchanged target error %v", err, wantErr)
				}

				if target.calls != 1 || target.method != tc.method || !reflect.DeepEqual(target.args, tc.args) {
					t.Fatalf("setter calls = %d, method = %q, args = %#v; want 1, %q, %#v",
						target.calls, target.method, target.args, tc.method, tc.args)
				}
			}
		})
	}
}
