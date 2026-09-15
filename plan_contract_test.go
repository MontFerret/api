package api_test

import "github.com/MontFerret/api"

// Metadata retrieval can fail without adding a context to Plan.Params.
var _ func(api.Plan) ([]string, error) = api.Plan.Params
