package pprof

import (
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/http/pprof"
	"strconv"

	"github.com/ccfish2/infra/pkg/hive/cell"
	"github.com/sirupsen/logrus"
	"github.com/spf13/pflag"
)

const (
	Prof        = "prof"
	ProfAddress = "prof-address"
	ProfPort    = "prof-port"
)

var Cell = cell.Module(
	"pprof-metrics",
	"Ppof metrics report",
	cell.Provide(newServer),
	cell.Invoke(func(s Server) {
	}),
)

type Config struct {
	Prof        bool
	ProfAddress string
	ProfPort    uint16
}

func (cfg Config) Flags(flags *pflag.FlagSet) {
	flags.Bool(Prof, cfg.Prof, "pprof")
	flags.String(ProfAddress, cfg.ProfAddress, "pprof addr")
	flags.Uint16(ProfPort, cfg.ProfPort, "pprof port")
}

type server struct {
	logger   logrus.FieldLogger
	addr     string
	port     uint16
	httpSrv  *http.Server
	listener net.Listener
}

// Start implements cell.HookInterface.
func (s *server) Start(cxt cell.HookContext) error {
	lis, err := net.Listen("tcp", net.JoinHostPort(s.addr, strconv.FormatUint(uint64(s.port), 10)))
	if err != nil {
		fmt.Printf("failed to announc listen %v", err)
		return err
	}
	s.listener = lis
	s.logger = s.logger.WithField("pprof http", "")

	mux := http.NewServeMux()
	mux.HandleFunc("/debug/pprof/", pprof.Index)
	mux.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
	mux.HandleFunc("/debug/pprof/profile", pprof.Profile)
	mux.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
	mux.HandleFunc("/debug/pprof/trace", pprof.Trace)
	s.httpSrv = &http.Server{Handler: mux}

	go func() {
		if err := s.httpSrv.Serve(s.listener); !errors.Is(err, http.ErrServerClosed) {
			s.logger.WithError(err)
		}
	}()
	s.logger.Info("starting pprof server")
	return nil
}

// Stop implements cell.HookInterface.
func (s *server) Stop(cxt cell.HookContext) error {
	err := s.httpSrv.Shutdown(cxt)
	return err
}

func newServer(lc cell.Lifecycle, log logrus.FieldLogger, cfg Config) Server {
	if !cfg.Prof {
		fmt.Println("does not enable pprof. skip")
		return nil
	}
	srv := &server{
		logger:  log,
		addr:    cfg.ProfAddress,
		port:    cfg.ProfPort,
		httpSrv: &http.Server{},
	}

	lc.Append(srv)
	return srv
}

type Server interface {
	Port() int
}

func (s *server) Port() int {
	return s.listener.Addr().(*net.TCPAddr).Port
}
