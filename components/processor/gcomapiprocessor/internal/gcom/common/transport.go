package common

import (
	"net"
	"net/http"
	"time"

	"github.com/mwitkow/go-conntrack"
)

// NewDefaultHTTPTransport returns the default http.Transport used but this client implementation.
func NewDefaultHTTPTransport() *http.Transport {
	transport := http.DefaultTransport.(*http.Transport).Clone()

	// Set both MaxIdleConns and MaxIdleConnsPerHost so that we keep an high number of open connections
	// even when the HTTP client is used to connect to a single host. For more information on why idle
	// connections may not work as you expect, please see: https://github.com/golang/go/issues/13801
	transport.MaxIdleConns = 100
	transport.MaxIdleConnsPerHost = 100

	return transport
}

// NewConntrackRoundTripper configures the Conntrack Dialer on the input http.Transport
// so we can instrument outbound connections
func NewConntrackRoundTripper(transport *http.Transport, name string) http.RoundTripper {
	transport.DialContext = conntrack.NewDialContextFunc(
		conntrack.DialWithTracing(),
		conntrack.DialWithName(name),
		conntrack.DialWithDialer(&net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 30 * time.Second,
		}),
	)
	return transport
}
