package controller

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/ccfish2/infra/pkg/option"
	"github.com/google/uuid"
)

var (
	globalStatus = NewManager()
)

type ControllerFunc func(ctx context.Context) error

type Group struct {
	Name string
}

type managedController struct {
	controller

	params       ControllerParams
	cancelDoFunc context.CancelFunc
}

type Manager struct {
	controllers controllerMap
	mutex       sync.RWMutex
}

func NewManager() *Manager {
	return &Manager{
		controllers: controllerMap{},
	}
}

func (m *Manager) removeAll() []*managedController {
	ctrls := []*managedController{}

	m.mutex.Lock()
	defer m.mutex.Unlock()

	if m.controllers == nil {
		return ctrls
	}
	for _, ctrl := range m.controllers {
		m.removeController(ctrl)
		ctrls = append(ctrls, ctrl)
	}

	return ctrls
}

func (c *managedController) stopController() {
	if c.cancelDoFunc != nil {
		c.cancelDoFunc()
	}
	close(c.stop)
}

func (m *Manager) removeController(ctrl *managedController) {
	ctrl.stopController()
	delete(m.controllers, ctrl.name)

	globalStatus.mutex.Lock()
	delete(globalStatus.controllers, ctrl.uuid)
	globalStatus.mutex.Unlock()

	fmt.Printf("Removed controller")
}

func (m *Manager) RemoveAll() {
	m.removeAll()
}

func (m *Manager) RemoveAllAndWait() {
	ctrls := m.removeAll()
	for _, ctrl := range ctrls {
		<-ctrl.terminated
	}
}

func (m *Manager) UpdateController(name string, params ControllerParams) {
	m.updateController(name, params)
}

func (c *managedController) updateParamsLocked(params ControllerParams) {
	if params.DoFunc == nil {
		params.DoFunc = func(ctx context.Context) error {
			return undefinedDoFunc(c.name)
		}
	}
	if params.StopFunc == nil {
		params.StopFunc = NoopFunc
	}

	maxInterval := time.Duration(option.Config.MaxControllerInterval) * time.Second
	if maxInterval > 0 && params.RunInterval > maxInterval {
		fmt.Printf("Limiting interval to %s", maxInterval)
		params.RunInterval = maxInterval
	}

	ctx := c.params.Context
	if c.params.CancelDoFuncOnUpdate && c.cancelDoFunc != nil {
		c.cancelDoFunc()
		c.params.Context = nil
	}

	if c.params.Context == nil {
		if params.Context == nil {
			ctx, c.cancelDoFunc = context.WithCancel(context.Background())
		} else {
			ctx, c.cancelDoFunc = context.WithCancel(params.Context)
		}
	}

	c.params = params
	c.params.Context = ctx
}

func (m *Manager) updateController(name string, params ControllerParams) *managedController {
	start := time.Now()

	m.mutex.Lock()
	defer m.mutex.Unlock()

	if m.controllers == nil {
		m.controllers = controllerMap{}
	}

	if params.Group.Name == "" {
		fmt.Printf(
			"Controller initialized with unpopulated group information. " +
				"Metrics will not be exported for this controller.")
	}

	ctrl, exists := m.controllers[name]
	if exists {
		fmt.Printf("Updating existing controller")
		ctrl.updateParamsLocked(params)

		// Notify the goroutine of the params update.
		select {
		case ctrl.update <- ctrl.params:
		default:
		}

		fmt.Println("Controller update time: ", time.Since(start))
	} else {
		return m.createControllerLocked(name, params)
	}

	return ctrl
}

func (m *Manager) createControllerLocked(name string, params ControllerParams) *managedController {
	ctrl := &managedController{
		controller: controller{
			name:       name,
			group:      params.Group,
			uuid:       uuid.New().String(),
			stop:       make(chan struct{}),
			update:     make(chan ControllerParams, 1),
			trigger:    make(chan struct{}, 1),
			terminated: make(chan struct{}),
		},
	}
	ctrl.updateParamsLocked(params)
	fmt.Printf("Starting new controller")

	m.controllers[ctrl.name] = ctrl

	globalStatus.mutex.Lock()
	globalStatus.controllers[ctrl.uuid] = ctrl
	globalStatus.mutex.Unlock()

	go ctrl.runController(ctrl.params)
	return ctrl
}
