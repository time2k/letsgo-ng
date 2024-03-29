package letsgo

import (
	"fmt"
	"log"
	"time"

	"github.com/gomodule/redigo/redis" //redigo
	"github.com/mna/redisc"            //redis cluster base on redigo
)

//Lredisc 结构体
type Lredisc struct {
	Redisc *redisc.Cluster
}

//newLredisc 返回一个Lredis结构体指针
func newLredisc() *Lredisc {
	return &Lredisc{}
}

//Init 连接redis cluster集群
func (c *Lredisc) Init(serverlist []string, options []redis.DialOption) error {
	c.Redisc = &redisc.Cluster{
		StartupNodes: serverlist,
		DialOptions:  options,
		CreatePool:   RedisCreatePool,
	}

	if err := c.Redisc.Refresh(); err != nil {
		return err
	}
	return nil
}

//GetConn 得到一个redis.Conn 如果Retry为true返回一个redisc retryConn否则是一个普通的redisc conn
func (c *Lredisc) GetConn(Retry bool) redis.Conn {
	rediscconn := c.Redisc.Get()
	if Retry == true {
		retryConn, err := redisc.RetryConn(rediscconn, 3, 100*time.Millisecond)
		if err != nil {
			log.Fatal("RetryConn failed:", err)
		}
		return retryConn
	} else {
		return rediscconn
	}
}

//DoOnce 映射redisc.Do 方法
func (c *Lredisc) DoOnce(commandName string, args ...interface{}) (reply interface{}, err error) {
	redisconn := c.Redisc.Get()
	defer redisconn.Close()
	if redisconn.Err() != nil {
		return nil, fmt.Errorf("Lredisc:err while conn: %s", redisconn.Err().Error())
	}
	return redisconn.Do(commandName, args...)
}

func (c *Lredisc) Close() error {
	return c.Redisc.Close()
}
