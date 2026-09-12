package main

import (
	"fmt"
	"net/http"

	"github.com/ccfish2/infra/pkg/hive/cell"
	"github.com/ccfish2/infra/pkg/stream"
)

var eventHandlerCell = cell.Module(
	"events-handler",
	"Implement event handler",
	cell.Provide(newEventsHandler),
)

func newEventsHandler(ee ExampleEvents) HTTPHandlerOptOut {
	return HTTPHandlerOptOut{
		HTTPHandler: HTTPHandler{
			Path: "/events",
			Handler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				if f, ok := w.(http.Flusher); ok {
					for ev := range stream.ToChannel[ExampleEvent](r.Context(), ee) {
						fmt.Fprintf(w, "%s\n", ev.Message)
						f.Flush()
					}
					return
				}
				for ev := range stream.ToChannel[ExampleEvent](r.Context(), ee) {
					fmt.Fprintf(w, "%s\n", ev.Message)
				}
			},
		},
	}
}
