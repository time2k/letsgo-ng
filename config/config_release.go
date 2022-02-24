// +build release

package config

import (
	"time"

	"github.com/afex/hystrix-go/hystrix"
)

var (
	//config缓存失效时间
	CONFIG_CACHE_EXPIRE int32 = 600

	//cachehttp相关设置
	CACHEHTTP_DIAL_TIMEOUT                 = time.Second * 2 //http dial 超时，包含DNS时间
	CACHEHTTP_RESPONSE_TIMEOUT             = time.Second * 2 //http响应超时
	CACHEHTTP_CHANNEL_BUFFER_LEN           = 100
	CACHEHTTP_DOWNGRADE_CACHE_EXPIRE int32 = 60
	CACHEHTTP_SELECT_TIMEOUT               = time.Second * 10

	//schedule相关设置
	SCHEDULE_CHANNEL_BUFFER_LEN = 100 //调度器channel buffer长度
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
