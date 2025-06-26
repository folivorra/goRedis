package transport

import "context"

type ToServe interface { // TODO: naming
	Start() error
	Shutdown(ctx context.Context) error
}
