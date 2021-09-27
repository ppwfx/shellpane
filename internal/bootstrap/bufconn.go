package bootstrap

import (
	"context"
	"net"

	"google.golang.org/grpc/test/bufconn"
)

type BufconnListener struct {
	*bufconn.Listener
}

func (l BufconnListener) Dial(network, addr string) (net.Conn, error) {
	return l.Listener.Dial()
}

func (l BufconnListener) DialContext(ctx context.Context, network, addr string) (net.Conn, error) {
	return l.Listener.Dial()
}
