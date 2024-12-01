package cell

import (
	"testing"
	"time"

	"github.com/ccfish2/infra/pkg/lock"
	"github.com/stretchr/testify/assert"
)

func createRoot(hr statusNodeReporter) (Scope, func()) {
	s := rootScope(FullModuleID{"root.foo"}, hr)
	s.start()
	return s, func() {
		flushAndClose(s, "test statussubreporter")
	}
}
func createMockReporter(assert assert.Assertions) (*mockReporter, func(), func(), func()) {
	l := new(lock.Mutex)
	last := new(Level)
	*last = StatusUnknown
	hr := &mockReporter{
		ok: func(s string) {
			l.Lock()
			defer l.Unlock()
			*last = StatusOK
		},
		stopped: func(s string) {
			l.Lock()
			defer l.Unlock()
			*last = StatusStopped
		},
		degraded: func(s string) {
			l.Lock()
			defer l.Unlock()
			*last = StatusDegraded
		},
	}

	okAssertion := func() {
		assert.Eventually(func() bool {
			l.Lock()
			defer l.Unlock()
			return *last == StatusOK
		}, 5*time.Second, 10*time.Millisecond)
	}

	degradedAssertion := func() {
		assert.Eventually(func() bool {
			l.Lock()
			defer l.Unlock()
			return *last == StatusDegraded
		}, 5*time.Second, 10*time.Millisecond)

	}
	stoppedAssertion := func() {
		assert.Eventually(func() bool {
			l.Lock()
			defer l.Unlock()
			return *last == StatusStopped
		}, 5*time.Second, 10*time.Millisecond)

	}
	return hr, okAssertion, degradedAssertion, stoppedAssertion
}

type mockReporter struct {
	ok, degraded, stopped func(string)
}

// setStatus implements statusNodeReporter.
func (m *mockReporter) setStatus(n Update) {
	switch n.Level() {
	case StatusOK:
		m.ok(n.String())
	case StatusDegraded:
		m.degraded(n.String())
	case StatusStopped:
		m.stopped(n.String())
	}
}

func Test_SubstatusReport(t *testing.T) {
	hr, assertok, degraded, _ := createMockReporter(*assert.New(t))
	s, done := createRoot(hr)
	defer done()

	fooScope := GetSubScope(s, "foo")
	GetHealthReporter(fooScope, "1").OK("ok")
	GetHealthReporter(fooScope, "2").OK("ok")
	GetHealthReporter(fooScope, "3").OK("ok")
	assertok()
	GetHealthReporter(fooScope, "1").Degraded("oops", nil)
	degraded()
	GetHealthReporter(fooScope, "1").Stopped("boom")
	assertok()

	foo2Sccope := GetSubScope(s, "foo2")
	GetHealthReporter(foo2Sccope, "1").OK("ok")
	assertok()
	bar := GetSubScope(s, "bar")
	GetHealthReporter(bar, "1").Degraded("boom", nil)
	degraded()
	GetHealthReporter(bar, "1").Stopped("Ok")
	assertok()
}
