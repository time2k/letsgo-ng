package config

import (
	"time"
)

//DBconfig 数据库配置结构体
type DBconfig struct {
	DBhostsip         string        //IP地址
	DBusername        string        //用户名
	DBpassword        string        //密码
	DBname            string        //表名
	DBcharset         string        //字符集
	DBconnMaxConns    int           //最大连接数
	DBconnMaxIdles    int           //最大空闲连接
	DBconnMaxLifeTime time.Duration //连接最大生命周期
}

//DBconfigStruct 多数据库并支持主从式配置结构体
type DBconfigStruct map[string]map[string]DBconfig

//RPCconfig rpc服务器配置结构体
type RPCconfig struct {
	Network          string //网络 可以是tcp udp或者http
	Address          string //具体地址及端口
	MicroserviceName string //微服务名称
}

const (
	StatusOk            int = 1 //ok
	StatusNoData        int = 2 //无数据
	StatusParamsNoValid int = 3 //参数错误
	StatusError         int = 4 //异常
)
