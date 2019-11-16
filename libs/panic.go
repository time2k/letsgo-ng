package Libs

import (
    "fmt"
	"runtime"
	"sync"
	"time"
)

//Mu panic使用的全局锁
var Mu sync.Mutex

//PanicFunc panic及recover函数
func PanicFunc() {
	if err := recover(); err != nil {
		Mu.Lock()
		now := time.Now()
		timet := now.Format("@2006-01-02 15:04:05")
		fmt.Printf("===Letsgo Panic Recover %s===\n%s\n",timet,err)
		PrintStack()
		fmt.Printf("===Letsgo Panic Recover End===\n\n")
		Mu.Unlock()
		
	}
}

//PrintStack 打印运行时堆栈信息
func PrintStack() {
    var buf [4096]byte
    n := runtime.Stack(buf[:], false)
    fmt.Printf("\n>>Stack Info<<\n%s\n", string(buf[:n]))
}