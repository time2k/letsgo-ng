package letsgo

import (
	"fmt"
	"time"

	"github.com/gomodule/redigo/redis"
	jsoniter "github.com/json-iterator/go"
)

//Cache 结构体
type Cache struct {
	Memcached           *Lmemcache
	Redis               Lrediser
	UseRedisOrMemcached int //使用哪种缓存 1-memcached 2-redis
}

//newCache 返回一个Cache结构体指针
func newCache() *Cache {
	return &Cache{}
}

//Init 初始化
func (c *Cache) Init() {
	if c.Memcached != nil {
		c.UseRedisOrMemcached = 1
	}
	if c.Redis != nil {
		c.UseRedisOrMemcached = 2
	}
}

//Get 获得缓存
func (c *Cache) Get(cachekey string, DataStruct interface{}) (bool, error) {
	var json = jsoniter.ConfigCompatibleWithStandardLibrary
	CacheGet := false
	var err error
	switch c.UseRedisOrMemcached {
	case 1:
		CacheGet, err = c.Memcached.Get(cachekey, DataStruct)
		if err != nil {
			return false, fmt.Errorf("[error]Cache Memcached get cache: %s", err.Error())
		}
	case 2:
		conn := c.Redis.GetConn(true)
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
			if err == redis.ErrNil {
				return false, nil
			}
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
	switch c.UseRedisOrMemcached {
	case 1:
		err := c.Memcached.Set(cachekey, DataStruct, expire)
		if err != nil {
			return fmt.Errorf("[error]Cache Memcached set cache: %s", err.Error())
		}
	case 2:
		conn := c.Redis.GetConn(true)
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
	switch c.UseRedisOrMemcached {
	case 1:
		err := c.Memcached.Delete(cachekey)
		if err != nil {
			return fmt.Errorf("[error]Cache Memcached delete cache: %s", err.Error())
		}
	case 2:
		conn := c.Redis.GetConn(true)
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

	switch c.UseRedisOrMemcached {
	case 1:
		return 1, nil
	case 2:
		conn := c.Redis.GetConn(true)
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

	switch c.UseRedisOrMemcached {
	case 1:
		return false, nil
	case 2:
		conn := c.Redis.GetConn(true)
		defer conn.Close()

		s, err := redis.ByteSlices(conn.Do("BRPOP", cachekey, timeout))
		if err != nil {
			if err != redis.ErrNil {
				return false, fmt.Errorf("[error]Cache Redisc BRPOP %s", err.Error())
			} else {
				return false, nil
			}
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

	switch c.UseRedisOrMemcached {
	case 1:
		return 0, nil
	case 2:
		conn := c.Redis.GetConn(true)
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
	switch c.UseRedisOrMemcached {
	case 1:
		return nil, fmt.Errorf("Memcached Don't support DO")
	case 2:
		conn := c.Redis.GetConn(true)
		defer conn.Close()

		return conn.Do(CMD, Params...)
	}
	return nil, nil
}

//PUB only for redis
func (c *Cache) PUB(channelname string, content string) error {
	switch c.UseRedisOrMemcached {
	case 1:
		return fmt.Errorf("Memcached Don't support PUB")
	case 2:
		conn := c.Redis.GetConn(true)
		defer conn.Close()

		if _, err := conn.Do("PUBLISH", channelname, content); err != nil {
			return err
		}
	}
	return nil
}

//GetSubConn only for redis
func (c *Cache) GETSUBCONN(channelname string) (*redis.PubSubConn, error) {
	switch c.UseRedisOrMemcached {
	case 1:
		return nil, fmt.Errorf("Memcached Don't support GETSUBCONN")
	case 2:
		conn := c.Redis.GetConn(false) //can't use redisc retry_conn
		//defer conn.Close()
		psc := redis.PubSubConn{Conn: conn}

		if err := psc.Subscribe(redis.Args{}.AddFlat(channelname)...); err != nil {
			return nil, err
		}
		return &psc, nil
	}
	return nil, nil
}

//SUB only for redis and use conn default timeout，you should use in "for" loop statment for sub
func (c *Cache) SUB(psc *redis.PubSubConn) (string, error) {
	switch c.UseRedisOrMemcached {
	case 1:
		return "", fmt.Errorf("Memcached Don't support SUB")
	case 2:
		switch n := psc.Receive().(type) {
		case error:
			return "", n
		case redis.Message:
			return string(n.Data), nil
		case redis.Subscription: //ignore subsciption message, recursive next
			return c.SUB(psc)
		}
	}
	return "", nil
}

//SUB only for redis and use timeout param，you should use in "for" loop statment for sub
func (c *Cache) TIMEOUTSUB(psc *redis.PubSubConn, timeout time.Duration) (string, error) {
	switch c.UseRedisOrMemcached {
	case 1:
		return "", fmt.Errorf("Memcached Don't support SUB")
	case 2:
		switch n := psc.ReceiveWithTimeout(timeout).(type) {
		case error:
			return "", n
		case redis.Message:
			return string(n.Data), nil
		case redis.Subscription: //ignore subsciption message, recursive next
			return c.SUB(psc)
		}
	}
	return "", nil
}

//Show 显示设置
func (c *Cache) Show() string {
	switch c.UseRedisOrMemcached {
	case 1:
		return "Cache use Memcached"
	case 2:
		return "Cache use Redis"
	}
	return ""
}
