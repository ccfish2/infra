package controller

import (
	"context"
	"errors"
	"fmt"
	"math"
	"sync"
	"time"

	"github.com/ccfish2/infra/pkg/hive/cell"
	"github.com/ccfish2/infra/pkg/inctimer"
	"github.com/sirupsen/logrus"
)

type controllerMap map[string]*managedController

type controller struct {
	// Constant after creation, safe to access without locking
	group  Group
	name   string
	uuid   string
	logger *logrus.Entry

	// Channels written to and/or closed by the manager
	stop    chan struct{}
	update  chan ControllerParams
	trigger chan struct{}

	// terminated is closed by the controller goroutine when it terminates
	terminated chan struct{}

	// Manipulated by the controller, read by the Manager, requires locking
	mutex             sync.RWMutex
	successCount      int
	lastSuccessStamp  time.Time
	failureCount      int
	consecutiveErrors int
	lastError         error
	lastErrorStamp    time.Time
	lastDuration      time.Duration
}

type ControllerParams struct {
	Group Group

	DoFunc                 ControllerFunc
	CancelDoFuncOnUpdate   bool
	StopFunc               ControllerFunc
	RunInterval            time.Duration
	MaxRetryInterval       time.Duration
	ErrorRetryBaseDuration time.Duration
	NoErrorRetry           bool
	Context                context.Context
	HealthReporter         cell.HealthReporter
}

func NewGroup(name string) Group {
	return Group{Name: name}
}

func undefinedDoFunc(name string) error {
	return fmt.Errorf("controller %s DoFunc is nil", name)
}

func NoopFunc(ctx context.Context) error {
	return nil
}

type ExitReason struct {
	error
}

func NewExitReason(reason string) ExitReason {
	return ExitReason{errors.New(reason)}
}

/*within kubernetes clusters, error is exaggerated */
func (c *controller) runController(params ControllerParams) {
	errorRetries := 1

	runTimer, timeDone := inctimer.New()
	defer timeDone()

	for {
		var err error

		interval := params.RunInterval

		start := time.Now()
		err = params.DoFunc(params.Context)
		duration := time.Since(start)

		c.mutex.Lock()
		c.lastDuration = duration
		fmt.Println("Controller func execution time: ", c.lastDuration)

		if err != nil {
			if params.Context.Err() != nil {
				err = NewExitReason("controller context canceled")
			}

			var exitReason ExitReason
			if errors.As(err, &exitReason) {
				c.recordSuccess(params.HealthReporter)
				c.lastError = exitReason

				fmt.Println("Controller run succeeded; waiting for next controller update or stop")
				interval = time.Duration(math.MaxInt64)
			} else {
				fmt.Printf("Controller run failed")
				c.recordError(err, params.HealthReporter)

				if !params.NoErrorRetry {
					if params.ErrorRetryBaseDuration != time.Duration(0) {
						interval = time.Duration(errorRetries) * params.ErrorRetryBaseDuration
					} else {
						interval = time.Duration(errorRetries) * time.Second
					}

					if params.MaxRetryInterval > 0 && interval > params.MaxRetryInterval {
						fmt.Printf("Cap retry interval to max allowed value")
						interval = params.MaxRetryInterval
					}

					errorRetries++
				}
			}
		} else {
			c.recordSuccess(params.HealthReporter)

			errorRetries = 1
			if interval == time.Duration(0) {
				fmt.Printf("Controller run succeeded; waiting for next controller update or stop")
				interval = time.Duration(math.MaxInt64)
			}
		}

		c.mutex.Unlock()

		select {
		case <-c.stop:
			goto shutdown
		case params = <-c.update:
		case <-runTimer.After(interval):
		case <-c.trigger:
		}
		select {
		case <-c.stop:
			goto shutdown
		default:
		}
	}

shutdown:
	fmt.Printf("Shutting down controller")
	if err := params.StopFunc(context.TODO()); err != nil {
		c.mutex.Lock()
		c.recordError(err, params.HealthReporter)
		c.mutex.Unlock()
		fmt.Printf("Error on Controller stop")
	}
	close(c.terminated)
}

func (c *controller) recordError(err error, hr cell.HealthReporter) {
	if hr != nil {
		hr.Degraded(c.name, err)
	}
	c.lastError = err
	c.lastErrorStamp = time.Now()
	c.failureCount++
	c.consecutiveErrors++

}

func (c *controller) recordSuccess(hr cell.HealthReporter) {
	if hr != nil {
		hr.OK(c.name)
	}

	c.lastError = nil
	c.lastSuccessStamp = time.Now()
	c.successCount++
	c.consecutiveErrors = 0

}
