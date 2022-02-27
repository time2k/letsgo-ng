package letsgo

import (
	"fmt"

	"github.com/gomodule/redigo/redis" //redigo
)

//Lredis 结构体
type Lredis struct {
	Redis *redis.Pool
}

//newLredis 返回一个Lredis结构体指针
func newLredis() *Lredis {
	return &Lredis{}
}

//Init 连接redis cluster集群
func (c *Lredis) Init(serverlist []string, options []redis.DialOption) error {
	var err error
	c.Redis, err = RedisCreatePool(serverlist[0], options...)

	if err != nil {
		return err
	}
	return nil
}

//GetConn 得到一个redis.Conn
func (c *Lredis) GetConn(Retry bool) redis.Conn {
	return c.Redis.Get()
}

//DoOnce 映射redisc.Do 方法
func (c *Lredis) DoOnce(commandName string, args ...interface{}) (reply interface{}, err error) {
	redisconn := c.Redis.Get()
	defer redisconn.Close()
	if redisconn.Err() != nil {
		return nil, fmt.Errorf("Lredis:err while conn: %s", redisconn.Err().Error())
	}
	return redisconn.Do(commandName, args...)
}

func (c *Lredis) Close() error {
	return c.Redis.Close()
}
