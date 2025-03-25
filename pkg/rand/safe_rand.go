package rand

import (
	"math/rand"

	"github.com/ccfish2/infra/pkg/lock"
)

/* SafeRand */
type SafeRand struct {
	mu lock.Mutex
	r  *rand.Rand
}

/*
NewSafeRand func
*/
func NewSafeRand(seed int64) *SafeRand {
	return &SafeRand{
		r: rand.New(rand.NewSource(seed)),
	}
}

func (sr *SafeRand) Seed(seed int64) {
	sr.mu.Lock()
	defer sr.mu.Unlock()
	sr.r.Seed(seed)
}

func (sr *SafeRand) Float64() float64 {
	sr.mu.Lock()
	defer sr.mu.Unlock()
	v := sr.r.Float64()
	return v
}

func (sr *SafeRand) Int63() int64 {
	sr.mu.Lock()
	defer sr.mu.Unlock()
	v := sr.r.Int63()
	return v
}

func (sr *SafeRand) Int63n(n int64) int64 {
	sr.mu.Lock()
	defer sr.mu.Unlock()
	v := sr.r.Int63n(n)
	return v
}

func (sr *SafeRand) Uint32(n int64) uint32 {
	sr.mu.Lock()
	v := sr.r.Uint32()
	sr.mu.Unlock()
	return v
}

func (sr *SafeRand) Uint64(n int64) uint64 {
	sr.mu.Lock()
	v := sr.r.Uint64()
	sr.mu.Unlock()
	return v
}

func (sr *SafeRand) Intn(n int) int {
	sr.mu.Lock()

	v := sr.r.Intn(n)
	sr.mu.Unlock()
	return v
}

func (sr *SafeRand) Perm(n int) []int {
	sr.mu.Lock()
	v := sr.r.Perm(n)
	sr.mu.Unlock()
	return v
}

func (sr *SafeRand) Shuffle(n int, swap func(i, j int)) {
	sr.mu.Lock()
	sr.r.Shuffle(n, swap)
	sr.mu.Unlock()
}
