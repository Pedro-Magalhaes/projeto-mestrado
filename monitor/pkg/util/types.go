package util

import "sync"

type ResourceState struct {
	CreatingWatcher bool
	BeeingWatched   bool
	KeepWorking     chan bool
}

type ResourceSafeMap struct {
	Mu          sync.Mutex
	ResourceMap map[string]*ResourceState
}

type SafeMap struct {
	mu sync.Mutex
	V  map[string]bool
}

type Runnable func(chan bool)

type FileChunkMsg struct {
	Msg    []byte `json:"msg"`
	Offset int64  `json:"offset"`
	Lenth  int    `json:"lenth"`
}

type InfoMsg struct {
	Path    string `json:"path"`
	Project string `json:"project"`
	Watch   bool   `json:"watch"`
}
