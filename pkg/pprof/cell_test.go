package pprof

import (
	"context"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/ccfish2/infra/pkg/hive"
	"github.com/ccfish2/infra/pkg/hive/cell"
	"go.uber.org/goleak"
)

func TestPProfSrv(t *testing.T) {
	defer goleak.VerifyNone(t)

	var testSrv Server
	h := hive.New(
		cell.Provide(newServer),
		cell.Config(Config{
			Prof:        false,
			ProfAddress: "localhost",
			ProfPort:    0,
		}),
		cell.Invoke(func(srv Server) {
			testSrv = srv
		}))
	if err := h.Start(context.Background()); err != nil {
		t.Fatalf("failed to start h, %v", err)
	}
	if testSrv != nil {
		t.Fatalf("listener unexpectedly started on port %d", testSrv.Port())
	}
	if err := h.Stop(context.Background()); err != nil {
		t.Fatalf("failed to stop %v", err)
	}
}

func TestHTTPHandler(t *testing.T) {
	defer goleak.VerifyNone(t)

	var testSrv Server
	h := hive.New(
		cell.Provide(newServer),
		cell.Config(Config{
			Prof:        true,
			ProfAddress: "localhost",
			ProfPort:    0,
		}),
		cell.Invoke(func(srv Server) {
			testSrv = srv
		}))
	if err := h.Start(context.Background()); err != nil {
		t.Fatalf("failed to start h, %v", err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, fmt.Sprintf("http://localhost:%d/debug/pprof/heap", testSrv.Port()), nil)
	if err != nil {
		t.Fatalf("failed to new http request %v", err)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("failed to execute req %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("failed to access the pprof heap endpoint %d", resp.StatusCode)
	}
	if err := h.Stop(ctx); err != nil {
		t.Fatalf("failed to stop the hive")
	}
}
