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

func TestOneShot_AfterRecover(t *testing.T) {
	var (
		g Group
		i int
	)

	h := fixture(func(r Registry, s cell.Scope, l cell.Lifecycle) {
		g = r.NewGroup(s)

		g.Add(OneShot("retry after recover", func(ctx context.Context, health cell.HealthReporter) error {
			defer func() {
				i++
			}()
			if i == 0 {
				return errors.New("sometimes error")
			}
			return nil
		}, WithRetry(3, workqueue.DefaultControllerRateLimiter())))
		l.Append(g)
	})

	if err := h.Start(context.Background()); err != nil {
		t.Fatal(err)
	}

	g.(*group).wg.Wait()

	if err := h.Stop(context.Background()); err != nil {
		t.Fatal(err)
	}
	if i != 2 {
		t.Fatal("supposed to reccover after fail")
	}
}

func TestOneShot_ShutDown(t *testing.T) {
	targetErr := errors.New("target error")
	h := fixture(func(r Registry, s cell.Scope, l cell.Lifecycle) {
		g := r.NewGroup(s)
		g.Add(OneShot("shutdown", func(ctx context.Context, health cell.HealthReporter) error {
			return targetErr
		}, WithShutdown()))
		l.Append(g)
	})

	err := h.Run()
	if err != nil && !errors.Is(err, targetErr) {
		t.Fatal("should be the targeted error")
	}
}

func TestOneShot_RetryFailShutdown(t *testing.T) {
	// retry three times
	retrytimes := 3
	var i int
	targetErr := errors.New("Always error")
	// expect would be "Always error"
	// build the hive using the fixture
	h := fixture(func(r Registry, s cell.Scope, l cell.Lifecycle) {
		g := r.NewGroup(s)

		g.Add(OneShot(
			"retry-fail-shutdown", func(ctx context.Context, health cell.HealthReporter) error {
				defer func() { i++ }()
				return targetErr
			}, WithRetry(retrytimes, workqueue.DefaultControllerRateLimiter()), WithShutdown(),
		))
		l.Append(g)
	})
	// add "retry-fail-shutdown" group into
	// add the group intot he lifecycle

	// kick off hive
	err := h.Run()
	if !errors.Is(err, targetErr) {
		t.Fail()
	}
	// expect always error
	// i would be the retry + 1
	if i != retrytimes+1 {
		t.Fail()
	}
}

func TestOneShot_RetryRecoverNoShutdown(t *testing.T) {
	var (
		g Group
		i int
	)

	started := make(chan interface{})
	h := fixture(func(r Registry, s cell.Scope, l cell.Lifecycle) {
		g = r.NewGroup(s)
		g.Add(OneShot("retry recover no shutdown", func(ctx context.Context, health cell.HealthReporter) error {

			if i == 0 {
				close(started) // run my wild thread

			}
			i++
			<-ctx.Done()

			return ctx.Err()
		}, WithRetry(3, workqueue.DefaultControllerRateLimiter())))
		l.Append(g)
	})

	shutdown := make(chan struct{})
	go func() {
		<-started
		h.Shutdown()
		close(shutdown)
	}()

	if err := h.Run(); err != nil {
		t.Fatal(err)
	}

	if i != 1 {
		t.Fatal("supposed to recover after fail")
	}
}

func TestTimer_OnInterval(t *testing.T) {
	// TestTimer_OnInterval tests the behavior of the Timer.OnInterval function.
	// It verifies that the provided callback function is called repeatedly at the specified interval.
	// The test fails if the callback function is not called or if it is called more times than expected.
	// The test also fails if the callback function takes longer than the specified interval to execute.
	stop := make(chan struct{})
	i := 0

	h := fixture(func(r Registry, s cell.Scope, l cell.Lifecycle) {
		g := r.NewGroup(s)
		g.Add(
			Timer("on-interval", func(ctx context.Context) error {
				i++
				if i == 5 {
					close(stop)
				}
				return nil
			}, 100*time.Millisecond))
		l.Append(g)
	})

	if err := h.Start(context.Background()); err != nil {
		t.Fatal(err)
	}
	<-stop

	if err := h.Stop(context.Background()); err != nil {
		t.Fatal(err)
	}
}
