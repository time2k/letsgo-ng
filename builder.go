package letsgo

/*
* DO NOT MODIFY
 */

import (
	"crypto/md5"
	"encoding/hex"
	"strconv"
	"time"

	"github.com/time2k/letsgo-ng/config"
)

//NewHTTPQueryBuilder 实例化
func NewHTTPQueryBuilder() *HTTPQueryBuilder {
	return &HTTPQueryBuilder{}
}

//NewDBQueryBuilder 实例化
func NewDBQueryBuilder() *DBQueryBuilder {
	return &DBQueryBuilder{}
}

//NewScheduleBuilder 实例化
func NewScheduleBuilder() *ScheduleBuilder {
	return &ScheduleBuilder{}
}

/*
* http builder define
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

//HTTPQueryBuilder 包含请求和多个响应channel的结构体
type HTTPQueryBuilder struct {
	Requests        []HTTPRequest
	ResponseCH      chan HTTPResponseResult
	CacheExpireTime int32
	DebugInfo
}

//GetBuilder 得到Builder信息
func (httpm *HTTPQueryBuilder) GetBuilder() *HTTPQueryBuilder {
	return httpm
}

//GetCacheExpire 得到缓存超时信息
func (httpm *HTTPQueryBuilder) GetCacheExpire() int32 {
	return httpm.CacheExpireTime
}

//GetHTTPUniqid 生成HTTP请求的唯一标识
func (httpm *HTTPQueryBuilder) GetHTTPUniqid(rtype string, method string, urls string, header map[string]string, postdata map[string]interface{}) string {
	var UniqBase string
	UniqBase += rtype + "|"
	UniqBase += method + "|"
	UniqBase += urls + "|"
	/*if header != nil {
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
	}*/
	UniqBase += strconv.FormatInt(time.Now().UnixNano(), 10)
	md5Ctx := md5.New()
	md5Ctx.Write([]byte(UniqBase))
	return hex.EncodeToString(md5Ctx.Sum(nil))
}

//SetRequest 设置HTTP请求
func (httpm *HTTPQueryBuilder) SetRequest(needcache bool, rtype string, method string, url string, header map[string]string, postdata map[string]interface{}) string {
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

//SetCacheExpire 设置缓存超时信息
func (httpm *HTTPQueryBuilder) SetCacheExpire(ExpireTime int32) {
	httpm.CacheExpireTime = ExpireTime
}

//InitHTTP 初始化http相关
func (httpm *HTTPQueryBuilder) InitHTTP() {
	httpm.ResponseCH = make(chan HTTPResponseResult, config.CACHEHTTP_CHANNEL_BUFFER_LEN)
}

//GetDebugInfo 得到debug信息
func (httpm *HTTPQueryBuilder) GetDebugInfo() *DebugInfo {
	return &httpm.DebugInfo
}

/*
* http model define end
 */

/*
* db builder define
 */

//DBQueryBuilder 数据库模块结构体
type DBQueryBuilder struct {
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
func (dbm *DBQueryBuilder) IsUseCache() bool {
	return dbm.UseCache
}

//GetCacheKey 得到缓存的key
func (dbm *DBQueryBuilder) GetCacheKey() string {
	return dbm.CacheKey
}

//SetCacheKey 设置缓存的key
func (dbm *DBQueryBuilder) SetCacheKey(CacheKey string) {
	dbm.CacheKey = CacheKey
}

//GetCacheExpire 得到缓存超时信息
func (dbm *DBQueryBuilder) GetCacheExpire() int32 {
	return dbm.CacheExpireTime
}

//SetCacheExpire 设置缓存超时信息
func (dbm *DBQueryBuilder) SetCacheExpire(ExpireTime int32) {
	dbm.CacheExpireTime = ExpireTime
}

//GetBuilder 得到Builder信息
func (dbm *DBQueryBuilder) GetBuilder() *DBQueryBuilder {
	return dbm
}

//SetSQL 设置SQL请求信息
func (dbm *DBQueryBuilder) SetSQL(sql string) {
	dbm.SQL = sql
}

//GetDbname 得到数据库名称
func (dbm *DBQueryBuilder) GetDbname() string {
	return dbm.DBName
}

//SetDbname 设置数据库名称
func (dbm *DBQueryBuilder) SetDbname(name string) {
	dbm.DBName = name
}

//SetSQLcondition 设置SQL查询条件
func (dbm *DBQueryBuilder) SetSQLcondition(con interface{}) {
	dbm.SQLcondition = append(dbm.SQLcondition, con)
}

//SetResult 设置接收的数据结构
func (dbm *DBQueryBuilder) SetResult(data interface{}) {
	dbm.Result = data
}

//GetDebugInfo 得到debug信息
func (dbm *DBQueryBuilder) GetDebugInfo() *DebugInfo {
	return &dbm.DebugInfo
}

/*
* db builder define end
 */

/*
* schedule builder define
 */

//FuncDesc 方法描述结构体
type FuncDesc struct {
	ModelFunc
	CommP CommonParams
	Args  []interface{}
}

//ScheduleBuilder 结构体
type ScheduleBuilder struct {
	FuncDescs []FuncDesc
	DataCH    chan ScheduleChan
	DebugInfo
}

//GetBuilder 得到Builder信息
func (schedulem *ScheduleBuilder) GetBuilder() *ScheduleBuilder {
	return schedulem
}

//SetSchedule 设置调度器信息
func (schedulem *ScheduleBuilder) SetSchedule(commp CommonParams, funcx ModelFunc, args ...interface{}) {
	schedulem.FuncDescs = append(schedulem.FuncDescs, FuncDesc{ModelFunc: funcx, CommP: commp, Args: args})
}

//GetDebugInfo 得到debug信息
func (schedulem *ScheduleBuilder) GetDebugInfo() *DebugInfo {
	return &schedulem.DebugInfo
}

//ScheduleChan 结构体
type ScheduleChan struct {
	SEQID string
	RET   BaseReturnData
}

//InitSchedule 初始化调度器相关
func (schedulem *ScheduleBuilder) InitSchedule() {
	schedulem.DataCH = make(chan ScheduleChan, config.SCHEDULE_CHANNEL_BUFFER_LEN)
}

/*
* schedule builder define end
 */
