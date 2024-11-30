package job

import (
	"context"
	"runtime"
	"testing"
	"time"

	"github.com/ccfish2/infra/pkg/hive"
	"github.com/ccfish2/infra/pkg/hive/cell"
	"github.com/ccfish2/infra/pkg/logging"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"go.uber.org/goleak"
)

var (
	tick    = 10 * time.Millisecond
	timeout = 5 * time.Second
)

func TestMain(m *testing.M) {
	cleanup := func(exitCode int) {
		// Force garbage-collection to force finalizers to run and catch
		// missing Event.Done() calls.
		runtime.GC()
	}
	goleak.VerifyTestMain(m, goleak.Cleanup(cleanup))
}

func fixture(fn func(Registry, cell.Scope, cell.Lifecycle)) *hive.Hive {
	logging.SetLogLevel(logrus.DebugLevel)
	return hive.New(
		Cell,
		cell.Module("test", "test module", cell.Invoke(fn)),
	)
}

// This test asserts that a OneShot jobs is started and completes. This test will timeout on failure
func TestOneShot_ShortRun(t *testing.T) {
	stop := make(chan struct{})

	h := fixture(func(r Registry, s cell.Scope, l cell.Lifecycle) {
		g := r.NewGroup(s)

		g.Add(
			OneShot("short", func(ctx context.Context, health cell.HealthReporter) error {
				defer close(stop)
				return nil
			}),
		)

		l.Append(g)
	})

	if assert.NoError(t, h.Start(context.Background())) {
		<-stop
		assert.NoError(t, h.Stop(context.Background()))
	}
}
