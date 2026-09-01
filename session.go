package api

import (
	"context"
	"io"
)

type Session interface {
	io.Closer
	Run(c context.Context) (Output, error)
}
