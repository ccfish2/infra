package rate

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func Test_Limiter(t *testing.T) {
	lim := NewLimiter(1*time.Second, 100)
	assert.NotNil(t, lim)

	b := lim.AllowN(100)
	assert.Equal(t, b, true)

	b = lim.Allow()
	assert.Equal(t, b, false)

	ctx, cancel := context.WithTimeout(context.Background(), 1500*time.Millisecond)
	err := lim.WaitN(ctx, 100)
	cancel()
	assert.Nil(t, err)

	// could not get space within 100 millisecond
	ctx, cancel = context.WithTimeout(context.Background(), 100*time.Millisecond)
	err = lim.Wait(ctx)
	cancel()
	assert.NotNil(t, err)

	// out of burst
	ctx, cancel = context.WithTimeout(context.Background(), 1500*time.Millisecond)
	err = lim.WaitN(ctx, 101)
	cancel()
	assert.NotNil(t, err)

	lim.Stop()
	defer func() {
		recover()
	}()
	lim.Allow()
}
