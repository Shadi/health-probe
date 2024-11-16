package probe

import (
	"crypto/tls"
	"net"
	"net/http"
	"net/http/httptrace"
	"time"

	"github.com/rs/zerolog/log"
)

type ReqResponse struct {
	Status    int
	Durations *ConnectionTimes
}
type ConnectionTimes struct {
	Dns          time.Duration
	Connect      time.Duration
	TlsHandshake time.Duration
	FirstByte    time.Duration
	Total        time.Duration
}

type TimedClient struct {
	client *http.Client
}

func NewClient(keepalive bool, maxConn int, hostConn int) *TimedClient {
	c := &TimedClient{
		client: httpClient(keepalive, maxConn, hostConn),
	}
	return c
}

// ToDo use config
func httpClient(keepalive bool, maxConn int, hostConn int) *http.Client {
	t := http.DefaultTransport.(*http.Transport).Clone()
	t.MaxIdleConns = maxConn
	t.MaxConnsPerHost = hostConn
	t.MaxIdleConnsPerHost = hostConn
	t.IdleConnTimeout = 5 * time.Second
	t.DisableKeepAlives = !keepalive
	t.DialContext = (&net.Dialer{Timeout: 5 * time.Second}).DialContext

	return &http.Client{
		Timeout:   time.Second * 5,
		Transport: t,
	}
}

func (c *TimedClient) Get(url string) ReqResponse {
	req, _ := http.NewRequest("GET", url, nil)

	var start, connect, dns, tlsHandshake time.Time
	cTime := &ConnectionTimes{}

	trace := &httptrace.ClientTrace{
		DNSStart: func(dsi httptrace.DNSStartInfo) { dns = time.Now() },
		DNSDone: func(ddi httptrace.DNSDoneInfo) {
			cTime.Dns = time.Since(dns)
		},

		TLSHandshakeStart: func() { tlsHandshake = time.Now() },
		TLSHandshakeDone: func(cs tls.ConnectionState, err error) {
			cTime.TlsHandshake = time.Since(tlsHandshake)
		},

		ConnectStart: func(network, addr string) { connect = time.Now() },
		ConnectDone: func(network, addr string, err error) {
			cTime.Connect = time.Since(connect)
		},

		GotFirstResponseByte: func() {
			cTime.FirstByte = time.Since(start)
		},
	}

	req = req.WithContext(httptrace.WithClientTrace(req.Context(), trace))
	start = time.Now()
	res, err := c.client.Transport.RoundTrip(req)
	if res != nil && res.Body != nil {
		defer res.Body.Close()
	} else {
		log.Debug().Str("url", url).Msg("Request returned no body")
	}
	cTime.Total = time.Since(start)
	if err != nil {
		log.Err(err).Str("url", url).Msg("Error calling service")
		return ReqResponse{Durations: nil, Status: 500}
	}
	return ReqResponse{Durations: cTime, Status: res.StatusCode}
}
