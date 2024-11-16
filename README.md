[![Build and Publish](https://github.com/shadi/health-probe/actions/workflows/build-and-publish.yaml/badge.svg)](https://github.com/shadi/health-probe/actions/workflows/build-and-publish.yaml)

# health-probe
Core part of [pingchck.co](https://pingchck.co/), Accepts a list of urls, then periodically send a request to each
and records response code and duration, and expose the results as Prometheus metrics.

List of urls can be passed using `--url`(short `-u`) flag at startup, you can also update the list
during runtime using the API `/urls`:`/urls?urls=https://pingchck.co&urls=https://s13h.xyz`


## API

* `/urls` endpoint to update urls list: `/urls?urls=https://pingchck.co&urls=https://s13h.xyz`
* `/add` add url to existing list during runtime: `/add?urls=https://pingchck.co`
* `/` to list current urls
* `/metrics` Prometheus metrics

## Metrics:
* `request_duration_seconds`: a gauge of Request durations, has 3 labels: `duration`, `response`, `url`.
* `request_duration_seconds{duration="total"}` returns the total response time of request.
  * `duration` label can have any of the following values:
  * `dns`: dns lookup time
  * `connect`: request connect duration
  * `tls`: tls handshake duration
  * `first`: time until first byte
  * `total`: Total request duration

* `response_code{response_code="200",url="https://pingchck.co"}` a counter of each url response type

##
Test using go:
`go run .`

The application has a number of configurable parameters, you can see them by using `-h` or `--help`:

`go run . --help`
