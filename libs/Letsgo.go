package Libs

import (
	"Letsgo2/Ltypedef"
	"database/sql"
	"log"
	"os"
	"sync"

	_ "github.com/go-sql-driver/mysql" //mysql
)

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
	Cache      *LCache
	JSONRPC    *LJsonrpcClient
	CacheQuery *LCacheQuery
	CacheHTTP  *LCacheHTTP
	Schedule   *LSchedule
	Logger     *log.Logger
	LoggerFile *os.File
	ModelPool  sync.Pool
	CacheLock  *LCacheLock
}

//CommonHeader 通用头
type CommonHeader struct {
	Status int    `json:"status"`
	Msg    string `json:"msg"`
	//ErrCode int
}

//CommonResp 通用返回
type CommonResp struct {
	CommonHeader `json:"header"`
	Body         interface{} `json:"body"`
}

//CommonModelParams 通用model参数
type CommonModelParams struct {
	*Letsgo
	Ltypedef.CommonReqParams
	IsDebug bool
}

//DebugModel debug信息
type DebugModel struct {
	DebugModelMU sync.Mutex `json:"-"`
	Info         []string   `json:"debug"`
}

//Add 向debug中添加信息
func (D *DebugModel) Add(info string) {
	D.DebugModelMU.Lock()
	D.Info = append(D.Info, info)
	D.DebugModelMU.Unlock()
}

//CommonRespWithDebug 带有debug信息的返回
type CommonRespWithDebug struct {
	CommonResp
	DebugModel
}

//NewLetsgo 返回一个Letsgo类型的结构体指针
func NewLetsgo() *Letsgo {
	return &Letsgo{}
}

//Init 初始化框架
func (L *Letsgo) Init() {
	//init modelpool
	L.ModelPool = sync.Pool{
		New: func() interface{} { return new(Lmodel) },
	}

	//init db
	L.DBC = make(map[string]DBSet)

	//传递自持
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
}

//MyLetsgo 框架自持变量
var MyLetsgo Letsgo
