package version

import (
	"sync"

	"github.com/blang/semver/v4"
)

// represents api-server
type ServerCapabilities struct {
	MinimalVersionMet bool
	EndpointSlice     bool
	EndpointSlicev1   bool
	LeaseResourceLock bool
}

type cachedVersion struct {
	mu           sync.RWMutex
	capabilities ServerCapabilities
	version      semver.Version
}

var (
	cached = cachedVersion{}
)

func Version() semver.Version {
	cached.mu.Lock()
	c := cached.version
	cached.mu.Unlock()
	return c
}

func Capabilities() ServerCapabilities {
	cached.mu.RLock()
	defer cached.mu.RUnlock()
	C := cached.capabilities
	return C
}
