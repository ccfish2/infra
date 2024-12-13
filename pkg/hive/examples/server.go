package main

import (
	"context"
	"net/http"

	"github.com/sirupsen/logrus"
	"github.com/spf13/pflag"

	"github.com/ccfish2/infra/pkg/hive"
	"github.com/ccfish2/infra/pkg/hive/cell"
)

var serverCell = cell.Module(
	"http-server",
	"Simple HTTP Server",

	cell.Config(defaultServerConfig),
	cell.Provide(newServer),
)

//
// Server API
//

type Server interface {
	ListenAddress() string
}

type serverConfig struct {
	ServerAddr string
}

func (cfg serverConfig) Flags(flags *pflag.FlagSet) {
	flags.String("server-listen-address", cfg.ServerAddr, "")
}

var defaultServerConfig = serverConfig{
	ServerAddr: ":8888",
}

type HTTPHandlerOptOut struct {
	cell.Out

	HTTPHandler HTTPHandler `group:"httphandlers,omitempty"`
}

type HTTPHandler struct {
	Path    string
	Handler http.HandlerFunc
}

type serverParams struct {
	cell.In

	Logger   logrus.FieldLogger
	Config   serverConfig
	LC       cell.Lifecycle
	Shutdown hive.Shutdowner
	handler  []HTTPHandler
}

type simpleServer struct {
	params serverParams
	server http.Server
}

func (s *simpleServer) ListenAddress() string {
	return s.server.Addr
}

func (s *simpleServer) listenAndServe() {
	s.params.Logger.WithField("server-addre", s.params.Config.ServerAddr).Info("Listening")
	err := s.server.ListenAndServe()
	if err != nil {
		s.server.Shutdown(context.Background())
		s.params.Logger.Error("http server could not be up.")
	}
}

func (s *simpleServer) Start(ctx cell.HookContext) error {
	go s.listenAndServe()
	return nil
}

func (s *simpleServer) Stop(ctx cell.HookContext) error {
	return s.server.Shutdown(ctx)
}

func newServer(params serverParams) Server {
	mux := http.NewServeMux()
	s := &simpleServer{params: params}
	s.server.Addr = s.params.Config.ServerAddr
	s.server.Handler = mux
	for _, handler := range s.params.handler {
		mux.HandleFunc(handler.Path, handler.Handler)
	}
	params.LC.Append(s)
	return s
}
