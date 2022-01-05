package letsgo

import (
	"database/sql"
	"log"
	"os"
	"time"

	"github.com/time2k/letsgo-ng/config"

	"github.com/afex/hystrix-go/hystrix"
	_ "github.com/go-sql-driver/mysql" //mysql
	"github.com/gomodule/redigo/redis"
	"github.com/labstack/echo"
)

//CommonParams Letsgo handler,model 通用参数
type CommonParams struct {
	//Letsgo      *Letsgo
	HTTPContext echo.Context
	Params      map[string]string
}

//Init 初始化
func (commp *CommonParams) Init() {
	commp.Params = make(map[string]string)
}

//SetParam 插入参数
func (commp *CommonParams) SetParam(name string, value string) bool {
	commp.Params[name] = value
	return true
}

//GetParam 读出参数
func (commp *CommonParams) GetParam(name string) string {
	v, ok := commp.Params[name]
	if ok != true {
		return ""
	}
	return v
}

//DBSet 支持1主1从的DBset
type DBSet struct {
	Master *sql.DB
	Slave  *sql.DB
}

//DBC DBSet集合
type DBC map[string]DBSet

//Letsgo 框架依赖功能结构体
type Letsgo struct {
	DBC
	Cache         *Cache
	CacheLock     *CacheLock
	DBQuery       *DBQuery
	HTTPQuery     *HTTPQuery
	JSONRPCClient *JSONRPCClient
	Schedules     *Schedule
	Logger        *log.Logger
	LoggerFile    *os.File
	ContextSet    contextSet
}

//NewLetsgo 返回一个Letsgo类型的结构体指针
func NewLetsgo() {
	Default = &Letsgo{}
}

//Init 初始化框架
func (L *Letsgo) Init() {
	//init ModulePool
	/*L.ModulePool = sync.Pool{
		New: func() interface{} { return new(Module) },
	}*/

	//hystrix setting
	hystrix.ConfigureCommand(config.HYSTRIX_DEFAULT_TAG, config.HYSTRIX_DEFAULT_CONFIG)
}

//InitDBQuery 初始化DBQuery
func (L *Letsgo) InitDBQuery(cfg config.DBconfigStruct) {
	//db init
	L.DBC = make(map[string]DBSet)
	L.DBQuery = newDBQuery()

	for k, v := range cfg {
		var err error
		DBset := DBSet{Master: nil, Slave: nil}

		//init Master
		DBset.Master, err = sql.Open("mysql", v["master"].DBusername+":"+v["master"].DBpassword+"@tcp("+v["master"].DBhostsip+")/"+v["master"].DBname+"?charset="+v["master"].DBcharset)
		if err != nil {
			log.Panicf("[error]Databases: connect error: %s", err.Error())
		}
		DBset.Master.SetMaxOpenConns(v["master"].DBconnMaxConns)
		DBset.Master.SetMaxIdleConns(v["master"].DBconnMaxIdles)
		DBset.Master.SetConnMaxLifetime(v["master"].DBconnMaxLifeTime)
		err = DBset.Master.Ping()
		if err != nil {
			log.Panicf("[error]Databases: connect error: %s", err.Error())
		}

		//init slave
		//判断是否是主从集群
		if v["slave"] != (config.DBconfig{}) {
			DBset.Slave, err = sql.Open("mysql", v["slave"].DBusername+":"+v["slave"].DBpassword+"@tcp("+v["slave"].DBhostsip+")/"+v["slave"].DBname+"?charset="+v["slave"].DBcharset)
			if err != nil {
				log.Panicf("[error]Databases: connect error: %s", err.Error())
			}
			DBset.Slave.SetMaxOpenConns(v["slave"].DBconnMaxConns)
			DBset.Slave.SetMaxIdleConns(v["slave"].DBconnMaxIdles)
			DBset.Master.SetConnMaxLifetime(v["slave"].DBconnMaxLifeTime)
			err = DBset.Slave.Ping()
			if err != nil {
				log.Panicf("[error]Databases: connect error: %s", err.Error())
			}
		} else {
			DBset.Slave = nil
		}

		//finally assign
		L.DBC[k] = DBset
	}

	L.DBQuery.SetDBset(L.DBC)
	L.DBQuery.SetCache(L.Cache)

}

//InitMemcached 初始化memcached
func (L *Letsgo) InitMemcached(MemcachedHost []string, MemcachedMaxIdleConns int, MemcachedMaxTimeout time.Duration) {
	//init cache
	L.Cache = newCache()
	//memcached
	L.Cache.Memcached = newLmemcache()
	L.Cache.Memcached.Conn(MemcachedHost...)
	L.Cache.Memcached.MaxIdleConns(MemcachedMaxIdleConns)
	L.Cache.Memcached.MaxTimeout(MemcachedMaxTimeout)

	L.Cache.Init()
}

//InitRedis 初始化redis
func (L *Letsgo) InitRedis(RedisClusterServer []string, RedisDialOption []redis.DialOption) {
	//init cache
	L.Cache = newCache()
	//redis
	L.Cache.Redisc = newLredisc()
	err := L.Cache.Redisc.Init(RedisClusterServer, RedisDialOption)
	if err != nil {
		log.Panicf("[error]RedisCluster: %s", err.Error())
	}

	L.Cache.Init()
}

//InitHTTPQuery 初始化http
func (L *Letsgo) InitHTTPQuery(HTTPLog string) {
	//init HTTPQuery
	L.HTTPQuery = newHTTPQuery()
	if L.Cache.UseRediscOrMemcached == 0 {
		log.Panicf("[error]HTTP use cache but cache doesn't init")
	}
	L.HTTPQuery.SetCache(L.Cache)
	L.HTTPQuery.Init(HTTPLog)
}

//InitLog 初始化日志
func (L *Letsgo) InitLog(LogFileName string) {
	logfile, err := os.OpenFile(LogFileName, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Panicf("Can't open the log file %s", LogFileName)
	}
	L.LoggerFile = logfile
	L.Logger = log.New(logfile, "[LetsGo Log]", log.Ldate|log.Ltime|log.Llongfile)
}

//InitSchedule 初始化并发器
func (L *Letsgo) InitSchedule() {
	L.Schedules = newSchedule()
	L.Schedules.Init()
}

//InitJSONRPC 初始化JSON RPC
func (L *Letsgo) InitJSONRPC(RPCConfig map[string]config.RPCconfig) {
	//init jsonrpc
	L.JSONRPCClient = NewJSONRPCClient()
	L.JSONRPCClient.Init()
	for service, rpcconfig := range RPCConfig {
		L.JSONRPCClient.Set(service, rpcconfig.Network, rpcconfig.Address)
	}
}

//InitMemConfig 初始化内存式配置
func (L *Letsgo) InitMemConfig() {
	//init config
	InitConfig()
}

//InitCacheLock 初始化缓存锁
func (L *Letsgo) InitCacheLock() {
	//init CacheLock
	L.CacheLock = newCacheLock()
	if L.Cache.UseRediscOrMemcached == 0 {
		log.Panicf("[error]CacheLock use cache but cache doesn't init")
	}
	L.CacheLock.Cache = L.Cache
}

//InitContextSet 初始化上下文集合
func (L *Letsgo) InitContextSet() {
	//init ContextSet
	L.ContextSet = newContextSet()
}

//Close 关闭Letsgo框架
func (L *Letsgo) Close() {
	for _, v := range L.DBC {
		v.Master.Close()
		//判断是否是主从集群
		if v.Slave != nil {
			v.Slave.Close()
		}
	}

	L.LoggerFile.Close()

	L.Cache.Redisc.Redisc.Close()

	L.ContextSet.CancelAll()
}

//Default 框架自持变量
var Default *Letsgo
