package api_test

import (
	"context"

	"github.com/MontFerret/api"
)

// Runtime version retrieval is fallible and context-aware.
var _ func(api.Runtime, context.Context) (api.Version, error) = api.Runtime.Version
