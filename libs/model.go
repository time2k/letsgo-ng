package Libs

/*
* DO NOT MODIFY
 */

import (
	"Letsgo2/Lconfig"
	"crypto/md5"
	"encoding/hex"
	"net/url"
)

//HTTPRequest http请求结构体
type HTTPRequest struct {
	UniqID    string
	NeedCache bool
	Rtype     string
	Method    string
	URL       string
	Header    map[string]string
	Postdata  map[string]string
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

//HTTPmodel 包含请求和多个响应channel的结构体
type HTTPmodel struct {
	Requests   []*HTTPRequest
	ResponseCH chan HTTPResponseResult
}

//SQLmodel SQL请求结构体
type SQLmodel struct {
	SQL          string
	SQLcondition []interface{}
	DataStruct   interface{}
	DbName       string
}

//ModelFunc 通用model描述，用作scheduler
type ModelFunc func(args ...interface{}) BaseReturnData

//FuncDesc 方法描述结构体
type FuncDesc struct {
	ModelFunc
	Args []interface{}
}

//ScheduleModel 结构体
type ScheduleModel struct {
	FuncDescs []FuncDesc
	DataCH    chan ScheduleChan
}

//ScheduleChan 结构体
type ScheduleChan struct {
	SEQID string
	RET   BaseReturnData
}

//Lmodel 框架model结构体
type Lmodel struct {
	CACHE_KEY         string
	CACHE_EXPIRE_TIME int32
	SQLmodel
	HTTPmodel
	ScheduleModel
	DebugModel
}

//NewLModel 返回一个Lmodel结构体指针
func NewLModel() *Lmodel {
	return new(Lmodel)
}

//GetCacheKey 得到缓存的key
func (c *Lmodel) GetCacheKey() string {
	return c.CACHE_KEY
}

//SetCacheKey 设置缓存的key
func (c *Lmodel) SetCacheKey(cache_key string) {
	c.CACHE_KEY = cache_key
}

//GetCacheExpire 得到缓存超时信息
func (c *Lmodel) GetCacheExpire() int32 {
	return c.CACHE_EXPIRE_TIME
}

//SetCacheExpire 设置缓存超时信息
func (c *Lmodel) SetCacheExpire(expire_time int32) {
	c.CACHE_EXPIRE_TIME = expire_time
}

//GetSQL 得到SQL请求信息
func (c *Lmodel) GetSQL() *SQLmodel {
	return &c.SQLmodel
}

//SetSQL 设置SQL请求信息
func (c *Lmodel) SetSQL(sql string) {
	c.SQLmodel.SQL = sql
}

//GetHTTP 得到HTTP请求信息
func (c *Lmodel) GetHTTP() *HTTPmodel {
	return &c.HTTPmodel
}

//GetSchedule 得到调度器信息
func (c *Lmodel) GetSchedule() *ScheduleModel {
	return &c.ScheduleModel
}

//SetSchedule 设置调度器信息
func (c *Lmodel) SetSchedule(funcx ModelFunc, args ...interface{}) {
	c.ScheduleModel.FuncDescs = append(c.ScheduleModel.FuncDescs, FuncDesc{ModelFunc: funcx, Args: args})
}

//GetDebugInfo 得到debug信息
func (c *Lmodel) GetDebugInfo() *DebugModel {
	return &c.DebugModel
}

//GetDbname 得到数据库名称
func (c *Lmodel) GetDbname() string {
	return c.SQLmodel.DbName
}

//SetDbname 设置数据库名称
func (c *Lmodel) SetDbname(name string) {
	c.SQLmodel.DbName = name
}

//SetSQLcondition 设置SQL查询条件
func (c *Lmodel) SetSQLcondition(con interface{}) {
	c.SQLmodel.SQLcondition = append(c.SQLmodel.SQLcondition, con)
}

//SetDataStruct 设置接收的数据结构
func (c *Lmodel) SetDataStruct(data interface{}) {
	c.SQLmodel.DataStruct = data
}

//GetHTTPUniqid 生成HTTP请求的唯一标识
func (c *Lmodel) GetHTTPUniqid(rtype string, method string, urls string, header map[string]string, postdata map[string]string) string {
	var uniq_base string
	uniq_base += rtype + "|"
	uniq_base += method + "|"
	uniq_base += urls + "|"
	if header != nil {
		data := make(url.Values)
		for k, v := range header {
			data[k] = []string{v}
		}
		uniq_base += data.Encode() + "|"
	}
	if postdata != nil {
		data := make(url.Values)
		for k, v := range postdata {
			data[k] = []string{v}
		}
		uniq_base += data.Encode() + "|"
	}
	md5Ctx := md5.New()
	md5Ctx.Write([]byte(uniq_base))
	return hex.EncodeToString(md5Ctx.Sum(nil))
}

//SetHTTP 设置HTTP请求
func (c *Lmodel) SetHTTP(needcache bool, rtype string, method string, url string, header map[string]string, postdata map[string]string) string {
	newRequest := new(HTTPRequest)
	newRequest.UniqID = c.GetHTTPUniqid(rtype, method, url, header, postdata)
	newRequest.NeedCache = needcache
	newRequest.Rtype = rtype
	newRequest.Method = method
	newRequest.URL = url
	newRequest.Header = header
	newRequest.Postdata = postdata

	c.HTTPmodel.Requests = append(c.HTTPmodel.Requests, newRequest)
	return newRequest.UniqID
}

//InitHTTP 初始化http相关
func (c *Lmodel) InitHTTP() {
	c.HTTPmodel.ResponseCH = make(chan HTTPResponseResult, Lconfig.CACHEHTTP_CHANNEL_BUFFER_LEN)
}

//InitSchedule 初始化调度器相关
func (c *Lmodel) InitSchedule() {
	c.ScheduleModel.DataCH = make(chan ScheduleChan, Lconfig.SCHEDULE_CHANNEL_BUFFER_LEN)
}
