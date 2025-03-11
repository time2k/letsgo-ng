package letsgo

/*
* DO NOT MODIFY
 */

import (
	"github.com/r3labs/sse/v2"
)

type SSE struct {
	ServerInstant *sse.Server
}

// 初始化sse服务端
func (s *SSE) Init() {
	s.ServerInstant = sse.New()        // create SSE broadcaster server
	s.ServerInstant.AutoReplay = false // do not replay messages for each new subscriber that connects
}

// sse客户端
func (s *SSE) NewClient(url string) *sse.Client {
	return sse.NewClient(url)
}
