package letsgo

import (
	"fmt"

	"github.com/gomodule/redigo/redis"
	jsoniter "github.com/json-iterator/go"
)

//Cache 结构体
type Cache struct {
	Memcached            *Lmemcache
	Redisc               *Lredisc
	UseRediscOrMemcached int //使用redisc或者memcached 1-memcached 2-redisc
}

//newCache 返回一个Cache结构体指针
func newCache() *Cache {
	return &Cache{}
}

//Init 初始化
func (c *Cache) Init() {
	if c.Memcached != nil {
		c.UseRediscOrMemcached = 1
	}
	if c.Redisc != nil {
		c.UseRediscOrMemcached = 2
	}
}

//SetMC 设置memcache连接
func (c *Cache) SetMC(mc *Lmemcache) {
	c.Memcached = mc
}

//SetRedis 设置memcache连接
func (c *Cache) SetRedis(redisc *Lredisc) {
	c.Redisc = redisc
}

//Get 获得缓存
func (c *Cache) Get(cachekey string, DataStruct interface{}) (bool, error) {
	var json = jsoniter.ConfigCompatibleWithStandardLibrary
	CacheGet := false
	var err error
	switch c.UseRediscOrMemcached {
	case 1:
		CacheGet, err = c.Memcached.Get(cachekey, DataStruct)
		if err != nil {
			return false, fmt.Errorf("[error]Cache Memcached get cache: %s", err.Error())
		}
	case 2:
		conn := c.Redisc.GetConn(true)
		defer conn.Close()

		isexist, err := redis.Int(conn.Do("EXISTS", cachekey))
		if err != nil {
			return false, fmt.Errorf("[error]Cache Redisc exists cmd: %s", err.Error())
		}
		if isexist == 0 {
			return false, nil
		}
		s, err := redis.String(conn.Do("GET", cachekey))
		if err != nil {
			return false, fmt.Errorf("[error]Cache Redisc get cache: %s", err.Error())
		}

		if s == "" {
			return false, fmt.Errorf("[error]Cache Redisc get cache empty")
		}

		if err := json.UnmarshalFromString(s, DataStruct); err != nil {
			return false, fmt.Errorf("[error]Cache Redisc get cache: %s", err.Error())
		}

		CacheGet = true
	}
	return CacheGet, nil
}

//Set 设置缓存
func (c *Cache) Set(cachekey string, DataStruct interface{}, expire int32) error {
	var json = jsoniter.ConfigCompatibleWithStandardLibrary
	switch c.UseRediscOrMemcached {
	case 1:
		err := c.Memcached.Set(cachekey, DataStruct, expire)
		if err != nil {
			return fmt.Errorf("[error]Cache Memcached set cache: %s", err.Error())
		}
	case 2:
		conn := c.Redisc.GetConn(true)
		defer conn.Close()

		str, err := json.MarshalToString(DataStruct)
		if err != nil {
			return fmt.Errorf("[error]Cache Redisc marshall struct: %s", err.Error())
		}

		_, err2 := conn.Do("SET", cachekey, str, "EX", expire)
		if err2 != nil {
			return fmt.Errorf("[error]Cache Redisc set cache: %s", err2.Error())
		}
	}
	return nil
}

//Delete 删除缓存
func (c *Cache) Delete(cachekey string) error {
	switch c.UseRediscOrMemcached {
	case 1:
		err := c.Memcached.Delete(cachekey)
		if err != nil {
			return fmt.Errorf("[error]Cache Memcached delete cache: %s", err.Error())
		}
	case 2:
		conn := c.Redisc.GetConn(true)
		defer conn.Close()

		_, err2 := conn.Do("DEL", cachekey)
		if err2 != nil {
			return fmt.Errorf("[error]Cache Redisc delete cache: %s", err2.Error())
		}

	}
	return nil
}

//SetNX only for redis distribut lock
func (c *Cache) SetNX(cachekey string, DataStruct interface{}, expire int32) (int, error) {
	var json = jsoniter.ConfigCompatibleWithStandardLibrary

	switch c.UseRediscOrMemcached {
	case 1:
		return 1, nil
	case 2:
		conn := c.Redisc.GetConn(true)
		defer conn.Close()

		str, err := json.MarshalToString(DataStruct)
		if err != nil {
			return 0, fmt.Errorf("[error]Cache Redisc marshall struct: %s", err.Error())
		}

		_, err2 := redis.String(conn.Do("SET", cachekey, str, "NX", "PX", expire))

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
func (c *Cache) BRPOP(cachekey string, DataStruct interface{}, timeout int32) (bool, error) {
	var json = jsoniter.ConfigCompatibleWithStandardLibrary

	switch c.UseRediscOrMemcached {
	case 1:
		return false, nil
	case 2:
		conn := c.Redisc.GetConn(true)
		defer conn.Close()

		s, err := redis.ByteSlices(conn.Do("BRPOP", cachekey, timeout))
		if err != nil {
			return false, fmt.Errorf("[error]Cache Redisc BRPOP %s", err.Error())
		}

		if err := json.UnmarshalFromString(string(s[1]), DataStruct); err != nil {
			return false, fmt.Errorf("[error]Cache Redisc BRPOP: %s", err.Error())
		}

		return true, nil
	}
	return false, nil
}

//LPUSH only for redis queue
func (c *Cache) LPUSH(cachekey string, DataStruct interface{}) (int, error) {
	var json = jsoniter.ConfigCompatibleWithStandardLibrary

	switch c.UseRediscOrMemcached {
	case 1:
		return 0, nil
	case 2:
		conn := c.Redisc.GetConn(true)
		defer conn.Close()

		str, err := json.MarshalToString(DataStruct)
		if err != nil {
			return 0, fmt.Errorf("[error]Cache Redisc marshall struct: %s", err.Error())
		}

		s, err := redis.Int(conn.Do("LPUSH", cachekey, str))
		if err != nil {
			return 0, fmt.Errorf("[error]Cache Redisc LPUSH %s", err.Error())
		}

		return s, nil
	}
	return 0, nil
}

//DO only for redis
func (c *Cache) DO(CMD string, Params ...interface{}) (interface{}, error) {
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
func (c *Cache) Show() string {
	switch c.UseRediscOrMemcached {
	case 1:
		return "Cache use Memcached"
	case 2:
		return "Cache use Redisc"
	}
	return ""
}
