package job

import (
	"context"
	"errors"
	"runtime"
	"testing"
	"time"

	"github.com/ccfish2/infra/pkg/hive"
	"github.com/ccfish2/infra/pkg/hive/cell"
	"github.com/ccfish2/infra/pkg/logging"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"go.uber.org/goleak"
	"k8s.io/client-go/util/workqueue"
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

func TestOneShot_LongRun(t *testing.T) {
	started, stopped := make(chan struct{}), make(chan struct{})

	h := fixture(func(r Registry, s cell.Scope, l cell.Lifecycle) {
		g := r.NewGroup(s)

		g.Add(OneShot("long", func(ctx context.Context, health cell.HealthReporter) error {
			close(started)
			<-ctx.Done()
			defer close(stopped)
			return nil
		}))
		l.Append(g)
	})
	if assert.NoError(t, h.Start(context.Background())) {
		<-started
		assert.NoError(t, h.Stop(context.Background()))
		<-stopped
	}
}

func TestOneShot_FailRetry(t *testing.T) {
	var (
		g Group
		i int
	)
	retries := 3
	h := fixture(func(r Registry, s cell.Scope, l cell.Lifecycle) {
		g = r.NewGroup(s)
		g.Add(OneShot("retries", func(ctx context.Context, health cell.HealthReporter) error {
			defer func() { i += 1 }()
			return errors.New("failed and retry")
		}, WithRetry(retries, workqueue.DefaultControllerRateLimiter())))

		l.Append(g)
	})
	if err := h.Start(context.Background()); err != nil {
		t.Fatal("failed to start", err)
	}

	g.(*group).wg.Wait()

	if err := h.Stop(context.Background()); err != nil {
		t.Fatal("failed to stop")
	}

	if i != retries+1 {
		t.Fatal("should have retired three times")
	}
}

func TestOneShot_RetryBackOff(t *testing.T) {
	ok := 0
	for i := 0; i < 5; i++ {
		failed, err := testOneShotFailBackOff()
		if err != nil {
			t.Fatal(err)
		}
		if !failed {
			ok++
		}
	}
	if ok == 0 {
		t.Fatal("0/5 test cases succeeded")
	}
}

func testOneShotFailBackOff() (bool, error) {
	var (
		g     Group
		i     int
		times []time.Time
	)
	const retries = 6
	failed := false
	h := fixture(func(r Registry, s cell.Scope, l cell.Lifecycle) {
		g = r.NewGroup(s)
		g.Add(OneShot("always back off", func(ctx context.Context, health cell.HealthReporter) error {
			defer func() {
				i++
			}()
			times = append(times, time.Now())
			return errors.New("always fail")
		}, WithRetry(retries, workqueue.NewItemExponentialFailureRateLimiter(50*time.Millisecond, 10*time.Second))))
	})
	if err := h.Start(context.Background()); err != nil {
		return true, err
	}
	g.(*group).wg.Wait()
	if err := h.Stop(context.Background()); err != nil {
		return true, err
	}
	var last time.Duration
	for i := 1; i < len(times); i++ {
		diff := times[i].Sub(times[i-1])
		if i > 2 {
			frat := uint64(diff * 10 / last * 10)
			if frat > 250 || frat < 150 {
				failed = true
			}
		}
		last = diff
	}
	return failed, nil
}
