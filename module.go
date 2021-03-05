package letsgo

/*
* DO NOT MODIFY
 */

import (
	"crypto/md5"
	"encoding/hex"
	"net/url"

	"github.com/time2k/letsgo-ng/config"
)

//Module 框架组件Module模块结构体
type Module struct {
	Database
	HTTPClient
	Schedule
}

//NewModule 返回一个Module结构体指针
func NewModule() *Module {
	return new(Module)
}

/*
* http module define
 */

//HTTPRequest http请求结构体
type HTTPRequest struct {
	UniqID    string
	NeedCache bool
	Rtype     string
	Method    string
	URL       string
	Header    map[string]string
	Postdata  map[string]interface{}
}

//HTTPResponseResult http响应结构体
type HTTPResponseResult struct {
	UniqID         string
	URL            string
	ResponseStatus int
	HTTPStatus     string
	HTTPStatusCode int
	ContentLength  int64
	Body           []byte
}

//HTTPClient 包含请求和多个响应channel的结构体
type HTTPClient struct {
	Requests        []HTTPRequest
	ResponseCH      chan HTTPResponseResult
	CacheExpireTime int32
	DebugInfo
}

//GetHTTP 得到HTTP请求信息
func (httpm HTTPClient) GetHTTP() *HTTPClient {
	return &httpm
}

//GetCacheExpire 得到缓存超时信息
func (httpm HTTPClient) GetCacheExpire() int32 {
	return httpm.CacheExpireTime
}

//GetHTTPUniqid 生成HTTP请求的唯一标识
func (httpm HTTPClient) GetHTTPUniqid(rtype string, method string, urls string, header map[string]string, postdata map[string]interface{}) string {
	var UniqBase string
	UniqBase += rtype + "|"
	UniqBase += method + "|"
	UniqBase += urls + "|"
	if header != nil {
		data := make(url.Values)
		for k, v := range header {
			data[k] = []string{v}
		}
		UniqBase += data.Encode() + "|"
	}
	if postdata != nil {
		data := make(url.Values)
		for k, v := range postdata {
			data[k] = []string{v.(string)}
		}
		UniqBase += data.Encode() + "|"
	}
	md5Ctx := md5.New()
	md5Ctx.Write([]byte(UniqBase))
	return hex.EncodeToString(md5Ctx.Sum(nil))
}

//SetHTTP 设置HTTP请求
func (httpm HTTPClient) SetHTTP(needcache bool, rtype string, method string, url string, header map[string]string, postdata map[string]interface{}) string {
	newRequest := HTTPRequest{}
	newRequest.UniqID = httpm.GetHTTPUniqid(rtype, method, url, header, postdata)
	newRequest.NeedCache = needcache
	newRequest.Rtype = rtype
	newRequest.Method = method
	newRequest.URL = url
	newRequest.Header = header
	newRequest.Postdata = postdata

	httpm.Requests = append(httpm.Requests, newRequest)
	return newRequest.UniqID
}

//InitHTTP 初始化http相关
func (httpm HTTPClient) InitHTTP() {
	httpm.ResponseCH = make(chan HTTPResponseResult, config.CACHEHTTP_CHANNEL_BUFFER_LEN)
}

//GetDebugInfo 得到debug信息
func (httpm HTTPClient) GetDebugInfo() *DebugInfo {
	return &httpm.DebugInfo
}

/*
* http model define end
 */

/*
* db module define
 */

//Database 数据库模块结构体
type Database struct {
	UseCache        bool
	CacheKey        string
	CacheExpireTime int32
	SQL             string
	SQLcondition    []interface{}
	Result          interface{}
	DBName          string
	DebugInfo
}

//IsUseCache 是否使用缓存
func (dbm Database) IsUseCache() bool {
	return dbm.UseCache
}

//GetCacheKey 得到缓存的key
func (dbm Database) GetCacheKey() string {
	return dbm.CacheKey
}

//SetCacheKey 设置缓存的key
func (dbm Database) SetCacheKey(CacheKey string) {
	dbm.CacheKey = CacheKey
}

//GetCacheExpire 得到缓存超时信息
func (dbm Database) GetCacheExpire() int32 {
	return dbm.CacheExpireTime
}

//SetCacheExpire 设置缓存超时信息
func (dbm Database) SetCacheExpire(ExpireTime int32) {
	dbm.CacheExpireTime = ExpireTime
}

//GetDB 得到SQL请求信息
func (dbm Database) GetDB() *Database {
	return &dbm
}

//SetSQL 设置SQL请求信息
func (dbm Database) SetSQL(sql string) {
	dbm.SQL = sql
}

//GetDbname 得到数据库名称
func (dbm Database) GetDbname() string {
	return dbm.DBName
}

//SetDbname 设置数据库名称
func (dbm Database) SetDbname(name string) {
	dbm.DBName = name
}

//SetSQLcondition 设置SQL查询条件
func (dbm Database) SetSQLcondition(con interface{}) {
	dbm.SQLcondition = append(dbm.SQLcondition, con)
}

//SetResult 设置接收的数据结构
func (dbm Database) SetResult(data interface{}) {
	dbm.Result = data
}

//GetDebugInfo 得到debug信息
func (dbm Database) GetDebugInfo() *DebugInfo {
	return &dbm.DebugInfo
}

/*
* db module define end
 */

/*
* schedule module define
 */

//FuncDesc 方法描述结构体
type FuncDesc struct {
	ModelFunc
	CommP CommonParams
	Args  []interface{}
}

//Schedule 结构体
type Schedule struct {
	FuncDescs []FuncDesc
	DataCH    chan ScheduleChan
	DebugInfo
}

//GetSchedule 得到调度器信息
func (schedulem Schedule) GetSchedule() *Schedule {
	return &schedulem
}

//SetSchedule 设置调度器信息
func (schedulem Schedule) SetSchedule(commp CommonParams, funcx ModelFunc, args ...interface{}) {
	schedulem.FuncDescs = append(schedulem.FuncDescs, FuncDesc{ModelFunc: funcx, CommP: commp, Args: args})
}

//GetDebugInfo 得到debug信息
func (schedulem Schedule) GetDebugInfo() *DebugInfo {
	return &schedulem.DebugInfo
}

//ScheduleChan 结构体
type ScheduleChan struct {
	SEQID string
	RET   BaseReturnData
}

//InitSchedule 初始化调度器相关
func (schedulem Schedule) InitSchedule() {
	schedulem.DataCH = make(chan ScheduleChan, config.SCHEDULE_CHANNEL_BUFFER_LEN)
}

/*
* schedule module define end
 */
