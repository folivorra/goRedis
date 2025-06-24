package transport

import "context"

type ToServe interface {
	Start() error
	Shutdown(ctx context.Context) error
}
