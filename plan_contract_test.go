package api_test

import (
	"context"

	"github.com/MontFerret/api"
)

// Parameter metadata retrieval is fallible and context-aware.
var _ func(api.Plan, context.Context) ([]string, error) = api.Plan.Params
