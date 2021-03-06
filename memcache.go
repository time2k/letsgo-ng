package letsgo

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"time"

	"github.com/bradfitz/gomemcache/memcache"
)

//Lmemcache 结构体
type Lmemcache struct {
	mc *memcache.Client
}

//newLmemcache 返回一个Lmemcache结构体指针
func newLmemcache() *Lmemcache {
	return &Lmemcache{}
}

//Conn 连接memcached服务器
func (c *Lmemcache) Conn(params ...string) {
	c.mc = memcache.New(params...)
}

//MaxIdleConns 设置memcached连接池的最大空闲连接，注意gomemcache不会主动释放空闲连接，请留出足够大此空闲连接
func (c *Lmemcache) MaxIdleConns(num int) {
	c.mc.MaxIdleConns = num
}

//MaxTimeout 设置memcached连接池的最长超时，在通讯中如果超时会有异常
func (c *Lmemcache) MaxTimeout(times time.Duration) {
	c.mc.Timeout = times
}

//Get memcached get方法
func (c *Lmemcache) Get(key string, stc interface{}) (bool, error) {
	if key == "" || stc == nil {
		return false, fmt.Errorf("[error]Memcache: Param invalid")
	}
	it, err := c.mc.Get(key)
	if err != nil { //cache miss or error
		if err != memcache.ErrCacheMiss {
			return false, fmt.Errorf("[error]Memcache get '%s': %s", key, err.Error())
		}

		return false, nil //special when cache miss
	}

	dec := gob.NewDecoder(bytes.NewReader(it.Value))
	err = dec.Decode(stc)
	if err != nil {
		return false, fmt.Errorf("[error]Memcache decode '%s': %s", key, err.Error())
	}
	return true, nil
}

//Set memcached set方法
func (c *Lmemcache) Set(key string, stc interface{}, expire int32) error {
	if key == "" || stc == nil {
		return fmt.Errorf("[error]Memcache Param invalid")
	}

	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	err := enc.Encode(stc)
	if err != nil {
		return fmt.Errorf("[error]Memcache encode '%s': %s", key, err.Error())
	}

	mcdata := &memcache.Item{
		Key:        key,
		Value:      buf.Bytes(),
		Expiration: expire,
	}

	err = c.mc.Set(mcdata)
	if err != nil {
		return fmt.Errorf("[error]Memcache set '%s': %s", key, err.Error())
	}
	return nil
}

//Delete memcached delete方法
func (c *Lmemcache) Delete(key string) error {
	if key == "" {
		return fmt.Errorf("[error]Memcache Param invalid")
	}

	return c.mc.Delete(key)
}
