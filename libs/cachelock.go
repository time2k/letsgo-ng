package Libs

import (
	"log"
	"math/rand"
	"strconv"
	"time"
)

//LCacheLock 结构体
type LCacheLock struct {
	Cache *LCache
}

//NewCacheLock 返回一个LCacheLock结构体指针
func NewCacheLock() *LCacheLock {
	return &LCacheLock{}
}

//Lock 上锁
func (c *LCacheLock) Lock(lockid int, prefix string, OWNER_ID string) {
	//必须使用redis
	if c.Cache.UseRediscOrMemcached == 1 {
		log.Println("CacheLock must use redis!")
		return
	}
	lockkey := "LOCK_" + prefix + "_" + strconv.Itoa(lockid)

	var LOCK_TIMEOUT int32 = 10000 //msec
	for {
		lock, _ := c.Cache.SetNX(lockkey, OWNER_ID, LOCK_TIMEOUT)
		if lock == 1 {
			//println("get lock by", OWNER_ID)
			return
		}
		//println("wait for release lock", OWNER_ID)
		//time.Sleep(time.Duration(10) * time.Millisecond)
		time.Sleep(time.Duration(rand.Intn(10)) * time.Millisecond)
	}
}

//Unlock 解锁
func (c *LCacheLock) Unlock(lockid int, prefix string, OWNER_ID string) {
	//必须使用redis
	if c.Cache.UseRediscOrMemcached == 1 {
		log.Println("CacheLock must use redis!")
		return
	}
	lockkey := "LOCK_" + prefix + "_" + strconv.Itoa(lockid)
	var lock_owner string
	isget, _ := c.Cache.Get(lockkey, &lock_owner)
	if isget == false {
		//println("lock missed!")
		return
	}
	if lock_owner == OWNER_ID {
		//println("release lock")
		c.Cache.Delete(lockkey)
	} else {
		//println("lock already change owner!")
	}
	return
}
