package probe

import (
	"fmt"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/rs/zerolog/log"
)

var (
	durationsGauge = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "request_duration_seconds",
		Help: "Total duration of request",
	}, []string{"url", "duration"})

	RespStatus = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "response_code",
		Help: "Count different responses",
	}, []string{"url", "response_code"})

	siteUp = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "response_status",
		Help: "Last response status",
	}, []string{"url", "response_code"})

	lastResp = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "last_response",
		Help: "Last response code",
	}, []string{"url"})
)

func init() {
	prometheus.MustRegister(durationsGauge)
	prometheus.MustRegister(RespStatus)
	prometheus.MustRegister(siteUp)
	prometheus.MustRegister(lastResp)
}

func TimeUrl(url string, c *TimedClient) {
	res := c.Get(url)
	printDurations(url, res.Durations)
	reportResponseStatus(url, res)
	reportDuration(url, res)
}

func printDurations(url string, durations *ConnectionTimes) {
	if durations == nil {
		log.Warn().Str("url", url).Msg("No Durations received")
		return
	}
	log.Info().
		Str("url", url).
		Dur("dns", durations.Dns).
		Dur("tls", durations.TlsHandshake).
		Dur("connect", durations.Connect).
		Dur("first_byte", durations.FirstByte).
		Dur("total", durations.Total).
		Msg("Durations")
}

func reportDuration(url string, resp ReqResponse) {
	if resp.Durations == nil {
		log.Warn().Str("url", url).Msg("No Durations received")
		return
	}
	durationsGauge.With(prometheus.Labels{"url": url, "duration": "dns"}).Set(resp.Durations.Dns.Seconds())
	durationsGauge.With(prometheus.Labels{"url": url, "duration": "tls"}).Set(resp.Durations.TlsHandshake.Seconds())
	durationsGauge.With(prometheus.Labels{"url": url, "duration": "connect"}).Set(resp.Durations.Connect.Seconds())
	durationsGauge.With(prometheus.Labels{"url": url, "duration": "first_byte"}).Set(resp.Durations.FirstByte.Seconds())
	durationsGauge.With(prometheus.Labels{"url": url, "duration": "total"}).Set(resp.Durations.Total.Seconds())
}

func reportResponseStatus(url string, resp ReqResponse) {
	RespStatus.With(prometheus.Labels{"url": url, "response_code": fmt.Sprintf("%v", resp.Status)}).Inc()
	lastResp.With(prometheus.Labels{"url": url}).Set(float64(resp.Status))
	if resp.Status >= 200 && resp.Status < 400 {
		siteUp.With(prometheus.Labels{"url": url, "response_code": fmt.Sprintf("%v", 200)}).Set(1)
	} else {
		siteUp.With(prometheus.Labels{"url": url, "response_code": fmt.Sprintf("%v", 200)}).Set(0)
	}
}
