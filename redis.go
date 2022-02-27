package letsgo

import (
	"time"

	"github.com/gomodule/redigo/redis" //redigo
	"github.com/time2k/letsgo-ng/config"
)

func RedisCreatePool(addr string, opts ...redis.DialOption) (*redis.Pool, error) {
	return &redis.Pool{
		MaxIdle:         config.REDIS_POOL_MAXIDLE,
		MaxActive:       config.REDIS_POOL_MAXACTIVE,
		IdleTimeout:     config.REDIS_POOL_IDLETIMEOUT,
		MaxConnLifetime: config.REDIS_POOL_MAXCONNLIFETIME,
		Wait:            config.REDIS_POLL_ALLOW_WAIT,
		Dial: func() (redis.Conn, error) {
			return redis.Dial("tcp", addr, opts...)
		},
		TestOnBorrow: func(c redis.Conn, t time.Time) error {
			_, err := c.Do("PING")
			return err
		},
	}, nil
}

//Lrediser interface
type Lrediser interface {
	Init(serverlist []string, options []redis.DialOption) error
	GetConn(Retry bool) redis.Conn
	DoOnce(commandName string, args ...interface{}) (reply interface{}, err error)
	Close() error
}
