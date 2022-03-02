package metadata

import (
	"context"
	"net"
)

const (
	key = "metadata"
)

type Metadata struct {
	LocalAddr  net.Addr
	RemoteAddr net.Addr
}

func NewContext(ctx context.Context, metadata Metadata) context.Context {
	return context.WithValue(ctx, key, metadata)
}

func FromContext(ctx context.Context) Metadata {
	return ctx.Value(key).(Metadata)
}
