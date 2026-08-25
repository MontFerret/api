package api

import (
	"context"
	"io"

	"github.com/MontFerret/api/result"
)

type Session interface {
	io.Closer
	Run(c context.Context) (result.Output, error)
}
