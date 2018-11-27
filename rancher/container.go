package rancher

import (
	"github.com/docker/libcompose/project"
	"golang.org/x/net/context"
)

type Container struct {
	id, name string
}

func NewContainer(id, name string) *Container {
	return &Container{
		id:   id,
		name: name,
	}
}

func (c *Container) ID() string {
	return c.id
}

func (c *Container) Name() string {
	return c.name
}

// TODO implement
func (c *Container) Port(ctx context.Context, port string) (string, error) {
	return "", project.ErrUnsupported
}

// TODO implement
func (c *Container) IsRunning(ctx context.Context) bool {
	return false
}
