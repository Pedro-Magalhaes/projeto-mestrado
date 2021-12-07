package util

type Runnable func(chan bool)

type FileChunkMsg struct {
	Msg    string `json:"msg"`
	Offset int64  `json:"offset"`
	Lenth  int    `json:"lenth"`
}

type InfoMsg struct {
	Path    string `json:"path"`
	Project string `json:"project"`
	Watch   bool   `json:"watch"`
}
