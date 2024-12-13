package main

import (
	"fmt"
	"net/http"

	"github.com/ccfish2/infra/pkg/hive/cell"
	"github.com/ccfish2/infra/pkg/stream"
)

var eventHandlerCell = cell.Module(
	"events handler",
	"Implement event handler",
	cell.Invoke(newEventsHandler),
)

func newEventsHandler(ee ExampleEvent) HTTPHandlerOptOut {
	return HTTPHandlerOptOut{
		HTTPHandler: HTTPHandler{
			Path: "/events",
			Handler: func(w http.ResponseWriter, r *http.Request) {
				f := w.(http.Flusher)
				w.WriteHeader(http.StatusOK)
				for ev := range stream.ToChannel[ExampleEvent](r.Context(), ee) {
					fmt.Fprintf(w, "%s", ev.Message)
					f.Flush()
				}
			},
		},
	}
}
