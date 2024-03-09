package http

import "context"

type ServerDependency struct {
	CreateUser func(ctx context.Context, username string) error
}
