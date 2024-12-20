package pprof

import (
	"errors"
	"fmt"
	"net"
	"net/http"
	"strconv"

	"github.com/ccfish2/infra/pkg/logging"
	"github.com/ccfish2/infra/pkg/logging/logfields"
)

var log = logging.DefaultLogger.WithField(logfields.LogSubsys, "pprof")

func Enable(addr string, port int) {
	apiAddr := net.JoinHostPort(addr, strconv.Itoa(port))
	go func() {
		if err := http.ListenAndServe(apiAddr, nil); !errors.Is(err, http.ErrServerClosed) {
			log.WithError(fmt.Errorf("failed to bump http server %v", err))
		}
	}()
	log.Info("succeed")
}
