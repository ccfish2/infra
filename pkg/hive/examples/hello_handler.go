package main

import (
	"net/http"

	"github.com/ccfish2/infra/pkg/hive/cell"
)

var helloHandler = cell.Module(
	"hellohandler",
	"Hello Handler",
	cell.Provide(newHelloHandler),
)

func newHelloHandler(exampleMetric exampleMetric) HTTPHandlerOptOut {
	return HTTPHandlerOptOut{
		HTTPHandler: HTTPHandler{
			Path: "hello",
			Handler: func(w http.ResponseWriter, r *http.Request) {
				exampleMetric.ExampleCounter.Inc()
				w.WriteHeader(http.StatusOK)
			},
		},
	}
}
