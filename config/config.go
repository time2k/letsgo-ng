// +build !debug
// +build !release

package config

import (
	"time"

	"github.com/afex/hystrix-go/hystrix"
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
	Network string //网络 可以是tcp udp或者http
	Address string //具体地址及端口
}

const (
	StatusOk            int = 1 //ok
	StatusNoData        int = 2 //无数据
	StatusParamsNoValid int = 3 //参数错误
	StatusError         int = 4 //异常
)

var (
	//config缓存失效时间
	CONFIG_CACHE_EXPIRE int32 = 600

	//cachehttp相关设置
	CACHEHTTP_DIAL_TIMEOUT                 = time.Second * 2 //http dial 超时，包含DNS时间
	CACHEHTTP_RESPONSE_TIMEOUT             = time.Second * 2 //http响应超时
	CACHEHTTP_CHANNEL_BUFFER_LEN           = 10
	CACHEHTTP_DOWNGRADE_CACHE_EXPIRE int32 = 60
	CACHEHTTP_SELECT_TIMEOUT               = time.Second * 10

	//schedule相关设置
	SCHEDULE_CHANNEL_BUFFER_LEN = 20 //调度器channel buffer长度
	SCHEDULE_MAX_CONCURRENT     = 20
	SCHEDULE_SELECT_TIMEOUT     = time.Second * 10

	//redis相关设置
	REDIS_POOL_MAXIDLE         = 10
	REDIS_POOL_MAXACTIVE       = 100
	REDIS_POOL_IDLETIMEOUT     = 5 * time.Minute
	REDIS_POOL_MAXCONNLIFETIME = 0 * time.Minute
	REDIS_POLL_ALLOW_WAIT      = true

	//hystrix相关设置
	HYSTRIX_DEFAULT_CONFIG hystrix.CommandConfig = hystrix.CommandConfig{
		Timeout:                3000,
		MaxConcurrentRequests:  1000,
		RequestVolumeThreshold: 1000,
		SleepWindow:            5000,
		ErrorPercentThreshold:  50,
	}

	HYSTRIX_DEFAULT_TAG string = "hystrix_default"
)
