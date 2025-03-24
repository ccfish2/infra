package rand

import (
	"math/rand"

	"github.com/ccfish2/infra/pkg/lock"
)

type SafeRand struct {
	mu lock.Mutex
	r  *rand.Rand
}

func NewSafeRand(seed int64) *SafeRand {
	return &SafeRand{
		r: rand.New(rand.NewSource(seed)),
	}
}

func (sr *SafeRand) Float64() float64 {
	sr.mu.Lock()
	defer sr.mu.Unlock()
	v := sr.r.Float64()
	return v
}
