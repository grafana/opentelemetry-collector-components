package common

import (
	"context"
	"net"
	"sync"

	"github.com/prometheus/client_golang/prometheus"
)

type CountingDialer struct {
	upstream func(ctx context.Context, network, addr string) (net.Conn, error)

	created prometheus.Counter
	closed  prometheus.Counter
}

func NewCountingDialer(
	upstream func(ctx context.Context, network string, addr string) (net.Conn, error),
	created, closed prometheus.Counter,
) *CountingDialer {
	return &CountingDialer{
		upstream: upstream,
		created:  created,
		closed:   closed,
	}
}

func (d *CountingDialer) Dial(ctx context.Context, network, addr string) (net.Conn, error) {
	c, err := d.upstream(ctx, network, addr)
	if err != nil {
		return c, err
	}

	d.created.Inc()
	return &conn{Conn: c, cleanup: d.closed.Inc}, nil
}

type conn struct {
	net.Conn

	once    sync.Once
	cleanup func()
}

func (c *conn) Close() error {
	c.once.Do(c.cleanup)
	return c.Conn.Close()
}
