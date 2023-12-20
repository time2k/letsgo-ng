package letsgo

import "sync"

/*
* DO NOT MODIFY
 */

/*
* debug info define
 */

//DebugInfo debug信息
type DebugInfo struct {
	lock sync.RWMutex
	Info []string `json:"debug"`
}

func NewDebugInfo() *DebugInfo {
	return &DebugInfo{Info: make([]string, 0)}
}

//Add 向debug中添加信息
func (D *DebugInfo) Add(info ...string) {
	D.lock.Lock()
	defer D.lock.Unlock()
	D.Info = append(D.Info, info...)
}

/*
* debug info define end
 */
