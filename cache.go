package letsgo

import (
	"fmt"

	"github.com/gomodule/redigo/redis"
	jsoniter "github.com/json-iterator/go"
)

type LCache struct {
	Memcached            *Lmemcache
	Redisc               *Lredisc
	UseRediscOrMemcached int //使用redisc或者memcached 1-memcached 2-redisc
}

//NewCache 返回一个LCache结构体指针
func NewLCache() *LCache {
	return &LCache{}
}

func (c *LCache) Init() {
	if c.Memcached != nil {
		c.UseRediscOrMemcached = 1
	}
	if c.Redisc != nil {
		c.UseRediscOrMemcached = 2
	}
}

//SetMC 设置memcache连接
func (c *LCache) SetMC(mc *Lmemcache) {
	c.Memcached = mc
}

//SetRedis 设置memcache连接
func (c *LCache) SetRedis(redisc *Lredisc) {
	c.Redisc = redisc
}

//Get 获得缓存
func (c *LCache) Get(cache_key string, DataStruct interface{}) (bool, error) {
	var json = jsoniter.ConfigCompatibleWithStandardLibrary
	CacheGet := false
	var err error
	switch c.UseRediscOrMemcached {
	case 1:
		CacheGet, err = c.Memcached.Get(cache_key, DataStruct)
		if err != nil {
			return false, fmt.Errorf("[error]LCache Memcached get cache: %s", err.Error())
		}
	case 2:
		conn := c.Redisc.GetConn(true)
		defer conn.Close()

		isexist, err := redis.Int(conn.Do("EXISTS", cache_key))
		if err != nil {
			return false, fmt.Errorf("[error]LCache Redisc exists cmd: %s", err.Error())
		}
		if isexist == 0 {
			return false, nil
		}
		s, err := redis.String(conn.Do("GET", cache_key))
		if err != nil {
			return false, fmt.Errorf("[error]LCache Redisc get cache: %s", err.Error())
		}

		if s == "" {
			return false, fmt.Errorf("[error]LCache Redisc get cache empty")
		}

		if err := json.UnmarshalFromString(s, DataStruct); err != nil {
			return false, fmt.Errorf("[error]LCache Redisc get cache: %s", err.Error())
		}

		CacheGet = true
	}
	return CacheGet, nil
}

//Set 设置缓存
func (c *LCache) Set(cache_key string, DataStruct interface{}, cache_expire int32) error {
	var json = jsoniter.ConfigCompatibleWithStandardLibrary
	switch c.UseRediscOrMemcached {
	case 1:
		err := c.Memcached.Set(cache_key, DataStruct, cache_expire)
		if err != nil {
			return fmt.Errorf("[error]LCache Memcached set cache: %s", err.Error())
		}
	case 2:
		conn := c.Redisc.GetConn(true)
		defer conn.Close()

		str, err := json.MarshalToString(DataStruct)
		if err != nil {
			return fmt.Errorf("[error]LCache Redisc marshall struct: %s", err.Error())
		}

		_, err2 := conn.Do("SET", cache_key, str, "EX", cache_expire)
		if err2 != nil {
			return fmt.Errorf("[error]LCache Redisc set cache: %s", err2.Error())
		}
	}
	return nil
}

//Delete 删除缓存
func (c *LCache) Delete(cache_key string) error {
	switch c.UseRediscOrMemcached {
	case 1:
		err := c.Memcached.Delete(cache_key)
		if err != nil {
			return fmt.Errorf("[error]LCache Memcached delete cache: %s", err.Error())
		}
	case 2:
		conn := c.Redisc.GetConn(true)
		defer conn.Close()

		_, err2 := conn.Do("DEL", cache_key)
		if err2 != nil {
			return fmt.Errorf("[error]LCache Redisc delete cache: %s", err2.Error())
		}

	}
	return nil
}

//SetNX only for redis distribut lock
func (c *LCache) SetNX(cache_key string, DataStruct interface{}, cache_expire int32) (int, error) {
	var json = jsoniter.ConfigCompatibleWithStandardLibrary

	switch c.UseRediscOrMemcached {
	case 1:
		return 1, nil
	case 2:
		conn := c.Redisc.GetConn(true)
		defer conn.Close()

		str, err := json.MarshalToString(DataStruct)
		if err != nil {
			return 0, fmt.Errorf("[error]LCache Redisc marshall struct: %s", err.Error())
		}

		_, err2 := redis.String(conn.Do("SET", cache_key, str, "NX", "PX", cache_expire))

		if err2 == redis.ErrNil {
			// The lock was not successful, it already exists.
			return 0, nil
		}
		if err2 != nil {
			return 0, err2
		}
		return 1, nil
	}
	return 0, nil
}

//BRPOP only for redis queue
func (c *LCache) BRPOP(cache_key string, DataStruct interface{}, timeout int32) (bool, error) {
	var json = jsoniter.ConfigCompatibleWithStandardLibrary

	switch c.UseRediscOrMemcached {
	case 1:
		return false, nil
	case 2:
		conn := c.Redisc.GetConn(true)
		defer conn.Close()

		s, err := redis.ByteSlices(conn.Do("BRPOP", cache_key, timeout))
		if err != nil {
			return false, fmt.Errorf("[error]LCache Redisc BRPOP %s", err.Error())
		}

		if err := json.UnmarshalFromString(string(s[1]), DataStruct); err != nil {
			return false, fmt.Errorf("[error]LCache Redisc BRPOP: %s", err.Error())
		}

		return true, nil
	}
	return false, nil
}

//LPUSH only for redis queue
func (c *LCache) LPUSH(cache_key string, DataStruct interface{}) (int, error) {
	var json = jsoniter.ConfigCompatibleWithStandardLibrary

	switch c.UseRediscOrMemcached {
	case 1:
		return 0, nil
	case 2:
		conn := c.Redisc.GetConn(true)
		defer conn.Close()

		str, err := json.MarshalToString(DataStruct)
		if err != nil {
			return 0, fmt.Errorf("[error]LCache Redisc marshall struct: %s", err.Error())
		}

		s, err := redis.Int(conn.Do("LPUSH", cache_key, str))
		if err != nil {
			return 0, fmt.Errorf("[error]LCache Redisc LPUSH %s", err.Error())
		}

		return s, nil
	}
	return 0, nil
}

//DO only for redis
func (c *LCache) DO(CMD string, Params ...interface{}) (interface{}, error) {
	switch c.UseRediscOrMemcached {
	case 1:
		return nil, fmt.Errorf("Memcached Don't support DO")
	case 2:
		conn := c.Redisc.GetConn(true)
		defer conn.Close()

		return conn.Do(CMD, Params...)
	}
	return nil, nil
}

//Show 显示设置
func (c *LCache) Show() string {
	switch c.UseRediscOrMemcached {
	case 1:
		return "LCache use Memcached"
	case 2:
		return "LCache use Redisc"
	}
	return ""
}
