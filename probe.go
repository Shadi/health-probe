package main

import (
	"fmt"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/shadi/health-probe/probe"

	"github.com/jessevdk/go-flags"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

var urls []string = make([]string, 0)
var urlsMutex sync.Mutex

const MAX_CONCURRENT_PROBES = 50

var limiter chan struct{} = make(chan struct{}, MAX_CONCURRENT_PROBES)

var opt struct {
	MetricsPort     int      `short:"m" long:"metrics_port" default:"9100" description:"Prometheus Metrics port"`
	ApiPort         int      `short:"p" long:"port" default:"8080" description:"Application API port for loading urls"`
	KeepAlive       bool     `short:"a" long:"keep_alive" description:"Keep connections alive between requests"`
	HostConnections int      `short:"c" long:"host_connections" default:"6" description:"Max connections per host"`
	ConnPoolSize    int      `short:"s" long:"pool_size" default:"200" description:"Connections pool size"`
	ProbeFreq       int      `short:"f" long:"freq" default:"30" description:"Probe frequency in seconds"`
	Urls            []string `short:"u" long:"url" default:"http://localhost:9100" description:"Default list of Urls to probe"`
}

func main() {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})

	parseFlags()

	ticker := time.NewTicker(time.Duration(opt.ProbeFreq) * time.Second)

	urlsListChan := make(chan []string)
	addUrlChan := make(chan string)

	for _, u := range opt.Urls {
		urls = append(urls, u)
	}

	apiHandler := probe.GetApiHandler(&urls, urlsListChan, addUrlChan)

	c := probe.NewClient(opt.KeepAlive, opt.ConnPoolSize, opt.HostConnections)

	http.Handle("/metrics", promhttp.Handler())
	go http.ListenAndServe(fmt.Sprintf(":%d", opt.MetricsPort), nil)

	go func() {
		s := &http.Server{
			Addr:         fmt.Sprintf(":%d", opt.ApiPort),
			Handler:      apiHandler.Router(),
			ReadTimeout:  5 * time.Second,
			WriteTimeout: 2 * time.Second,
		}
		err := s.ListenAndServe()
		if err != nil {
			log.Fatal().Err(err).Msg("Detected fatal error")
		}
	}()

	quit := make(chan struct{})

	for {
		select {
		case <-ticker.C:
			go probeUrls(c)
		case u := <-urlsListChan:
			updateUrls(u)
		case s := <-addUrlChan:
			addUrl(s)
		case <-quit:
			ticker.Stop()
			return
		}
	}
}

func parseFlags() {
	_, e := flags.Parse(&opt)
	if e != nil {
		log.Fatal().Err(e).Msg("Error parsing arguments")
	}
	log.Debug().Msgf("received params,p:%d,m:%d,u:%s",
		opt.ApiPort, opt.MetricsPort, opt.Urls)

}

func probeUrls(c *probe.TimedClient) {
	urlsMutex.Lock()
	defer urlsMutex.Unlock()
	for _, u := range urls {
		limiter <- struct{}{}
		go func(url string) {
			probe.TimeUrl(url, c)
			<-limiter
		}(u)
	}
}

func updateUrls(us []string) {
	log.Info().Msgf("Received update request with %v", us)
	urlsMutex.Lock()
	defer urlsMutex.Unlock()
	urls = nil
	urls = append(urls, us...)
}

func addUrl(s string) {
	log.Info().Msgf("Adding a url %s", s)
	urlsMutex.Lock()
	defer urlsMutex.Unlock()
	urls = append(urls, s)
}
