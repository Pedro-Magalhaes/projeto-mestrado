package util

import "sync"

type ResourceState struct {
	CreatingWatcher bool
	BeeingWatched   bool
}

type ResourceSafeMap struct {
	Mu          sync.Mutex
	ResourceMap map[string]ResourceState
}

type SafeMap struct {
	mu sync.Mutex
	V  map[string]bool
}

type Runnable func(chan bool)
