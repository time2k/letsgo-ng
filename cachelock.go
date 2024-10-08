package letsgo

import (
	"log"
	"math/rand"
	"strconv"
	"time"
)

// CacheLock 结构体
type CacheLock struct {
	Cache *Cache
}

// newCacheLock 返回一个CacheLock结构体指针
func newCacheLock() *CacheLock {
	return &CacheLock{}
}

// Lock 上锁
func (c *CacheLock) Lock(lockid int, prefix string, OWNER string, expiremilseconds int) {
	//必须使用redis
	if c.Cache.UseRedisOrMemcached == 1 {
		log.Println("CacheLock must use redis!")
		return
	}
	lockkey := "LOCK_" + prefix + "_" + strconv.Itoa(lockid)

	var LOCK_TIMEOUT int32 = 0
	if expiremilseconds == 0 {
		LOCK_TIMEOUT = 10000 //msec
	} else {
		LOCK_TIMEOUT = int32(expiremilseconds)
	}
	for {
		lock, _ := c.Cache.SetNX(lockkey, OWNER, LOCK_TIMEOUT)
		if lock == 1 {
			//println("get lock by", OWNER_ID)
			return
		}
		//println("wait for release lock", OWNER_ID)
		//time.Sleep(time.Duration(10) * time.Millisecond)
		time.Sleep(time.Duration(rand.Intn(10)) * time.Millisecond)
	}
}

// Unlock 解锁
func (c *CacheLock) Unlock(lockid int, prefix string, OWNER string) {
	//必须使用redis
	if c.Cache.UseRedisOrMemcached == 1 {
		log.Println("CacheLock must use redis!")
		return
	}
	lockkey := "LOCK_" + prefix + "_" + strconv.Itoa(lockid)
	var lockowner string
	isget, _ := c.Cache.Get(lockkey, &lockowner)
	if isget == false {
		//println("lock missed!")
		return
	}
	if lockowner == OWNER {
		//println("release lock")
		c.Cache.Delete(lockkey)
	} else {
		//println("lock already change owner!")
	}
	return
}
