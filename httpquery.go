package letsgo

import (
	"bytes"
	"compress/gzip"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"math/rand"
	"mime/multipart"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/time2k/letsgo-ng/config"

	"github.com/afex/hystrix-go/hystrix"
	"github.com/bitly/go-simplejson"
)

// HTTPQueryer 为cachehttp的接口
type HTTPQueryer interface {
	GetCacheExpire() int32
	GetBuilder() *HTTPQueryBuilder
	GetDebugInfo() *DebugInfo
	InitHTTP()
}

// HTTPQuery 为cachehttp的结构体
type HTTPQuery struct {
	CL              *http.Client
	Cache           *Cache
	HTTPcounter     int
	HTTPcounterLock sync.Mutex
	Logfile         string
}

// newHTTPQuery 返回一个HTTPQuery的结构体指针
func newHTTPQuery() *HTTPQuery {
	return &HTTPQuery{}
}

// Init 初始化HTTPQuery结构体
func (c *HTTPQuery) Init(logfile string) {
	c.CL = &http.Client{
		Transport: &http.Transport{
			Dial: func(netw, addr string) (net.Conn, error) {
				c, err := net.DialTimeout(netw, addr, time.Duration(time.Second*config.CACHEHTTP_DIAL_TIMEOUT))
				if err != nil {
					return nil, err
				}
				return c, nil

			},
			//MaxIdleConnsPerHost:   10,
			DisableKeepAlives:     true,
			ResponseHeaderTimeout: time.Second * config.CACHEHTTP_RESPONSE_TIMEOUT,
		},
	}
	c.Logfile = logfile
}

// SetCache 设置cache
func (c *HTTPQuery) SetCache(cache *Cache) {
	c.Cache = cache
}

// AddCounter 内置计数器++
func (c *HTTPQuery) AddCounter() {
	c.HTTPcounterLock.Lock()
	defer c.HTTPcounterLock.Unlock()
	c.HTTPcounter++
}

// SubCounter 内置计数器--
func (c *HTTPQuery) SubCounter() {
	c.HTTPcounterLock.Lock()
	defer c.HTTPcounterLock.Unlock()
	c.HTTPcounter--
}

// RandNum 在指定范围随机输出数字
func RandNum(ran int) int {
	t := time.Now().UnixNano()
	rand.Seed(t)
	rd := rand.Intn(ran) //[0,n)
	return rd
}

// GenUniqID 生成与时间有关的随机字符
func (c *HTTPQuery) GenUniqID() string {
	un := time.Now().UnixNano()
	md5Ctx := md5.New()
	md5Ctx.Write([]byte(strconv.FormatInt(un, 10) + strconv.Itoa(RandNum(1000))))
	cipherStr := hex.EncodeToString(md5Ctx.Sum(nil))
	return cipherStr
}

// SampleHTTPQuery 定义client
func (c *HTTPQuery) SampleHTTPQuery(rq HTTPRequest, debug *DebugInfo, ret chan HTTPResponseResult) error {
	//如果http请求响应日志有定义
	httpLog := new(log.Logger)
	if c.Logfile != "" {
		//定义一个文件
		logFile, err := os.OpenFile(c.Logfile, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
		defer logFile.Close()
		if err != nil {
			log.Fatalln("open file error !")
		}
		httpLog = log.New(logFile, "\n[Letsgo-cacheHTTP] ", log.LstdFlags)
	}

	var pbody io.Reader
	httpRes := HTTPResponseResult{}
	httpRes.UniqID = rq.UniqID
	httpRes.URL = rq.URL

	if rq.Postdata != nil {
		if _, ok := rq.Header["Content-Type"]; !ok { //默认类型为application/x-www-form-urlencoded
			data := make(url.Values)
			for k, v := range rq.Postdata {
				data.Add(k, v.(string))
			}
			pbody = strings.NewReader(data.Encode())
		} else if rq.Header["Content-Type"] == "application/x-www-form-urlencoded" {
			data := make(url.Values)
			for k, v := range rq.Postdata {
				data.Add(k, v.(string))
			}
			pbody = strings.NewReader(data.Encode())
		} else if rq.Header["Content-Type"] == "multipart/form-data" {
			var b bytes.Buffer
			w := multipart.NewWriter(&b)
			for k, v := range rq.Postdata {
				if strings.Contains(k, "file:") { //如果key中包含file:，则认为是文件型
					strsplit := strings.Split(k, ":")
					filename := strsplit[1]
					fw, err := w.CreateFormFile(filename, "file")
					if err != nil {
						debug.Add(fmt.Sprintf("CacheHTTP multipart create error: %s", err.Error()))
						log.Println("[error]CacheHTTP multipart create error:", err.Error())
						return err
					}
					file := v.(*multipart.FileHeader)
					fs, err := file.Open()
					defer fs.Close()
					if err != nil {
						debug.Add(fmt.Sprintf("CacheHTTP multipart open error: %s", err.Error()))
						log.Println("[error]CacheHTTP multipart open error:", err.Error())
						return err
					}
					if _, err := io.Copy(fw, fs); err != nil {
						debug.Add(fmt.Sprintf("CacheHTTP multipart io.Copy error: %s", err.Error()))
						log.Println("[error]CacheHTTP multipart io.Copy error:", err.Error())
						return err
					}
				} else {
					fw, err := w.CreateFormField(k)
					if err != nil {
						debug.Add(fmt.Sprintf("CacheHTTP multipart create error: %s", err.Error()))
						log.Println("[error]CacheHTTP multipart create error:", err.Error())
						return err
					}
					if _, err = fw.Write([]byte(v.(string))); err != nil {
						debug.Add(fmt.Sprintf("CacheHTTP multipart write error: %s", err.Error()))
						log.Println("[error]CacheHTTP multipart write error:", err.Error())
						return err
					}
				}
			}
			w.Close()
			pbody = &b
			//set header
			rq.Header["Content-Type"] = w.FormDataContentType()
		} else { //json body, binary body etc
			pbody = strings.NewReader(rq.Postdata["0"].(string))
		}
	}

	req, err := http.NewRequest(rq.Method, rq.URL, pbody)
	if err != nil {
		log.Println("[error]CacheHTTP gen newRequest:", err.Error())
	}
	//增加header
	req.Header.Add("User-Agent", "Mozilla/5.0")
	req.Header.Add("Accept-Encoding", "deflate")

	//增加默认类型头 application/x-www-form-urlencoded
	if (rq.Method == "POST" || rq.Method == "post") && rq.Postdata != nil {
		if _, ok := rq.Header["Content-Type"]; !ok {
			req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		}
	}

	for k, v := range rq.Header {
		req.Header.Set(k, v)
	}

	debug.Add(fmt.Sprintf("Send HTTP Query: %s", rq.URL))
	start := time.Now()

	//原始http请求
	if c.Logfile != "" {
		requestDump, err := httputil.DumpRequest(req, true)
		if err != nil {
			fmt.Println(err)
		}
		httpLog.Println("\n---Request---\n" + string(requestDump))
	}

	//解析url，获得请求的域名和参数
	/*request, err := url.Parse(rq.URL)
	if err != nil {
		panic(err)
	}*/

	//降级标签采用host/path形式进行定义，比如说：geo.mob.app.letv.com/geo
	//hystrix_tag := request.Host + request.Path
	hystrixtag := config.HYSTRIX_DEFAULT_TAG
	// params, _ := url.ParseQuery(request.RawQuery)

	hystrix.Do(hystrixtag, func() error {
		resp, err := c.CL.Do(req)
		end := time.Since(start)
		debug.Add(fmt.Sprintf("HTTP Time Cost: %f ms", end.Seconds()*1000))
		if err != nil {
			//不抛出错误而是接口降级
			debug.Add(fmt.Sprintf("HTTP Query Downgrade: %s", err.Error()))
			log.Println("[error]CacheHTTP request error:", err.Error())
			return err
		}
		if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusAccepted && resp.StatusCode != http.StatusNoContent {
			// //不抛出错误而是接口降级
			debug.Add(fmt.Sprintf("HTTP Query Downgrade: non-200 StatusCode:%s", rq.URL))
			log.Println("[error]CacheHTTP request got non-200 StatusCode:", rq.URL)

			httpRes.ResponseStatus = -1
		} else {
			httpRes.ResponseStatus = 0
		}
		debug.Add(fmt.Sprintf("HTTP Query Result: status:%s, content length:%d, url:%s", resp.Status, resp.ContentLength, rq.URL))
		defer resp.Body.Close()

		//原始http响应体
		if c.Logfile != "" {
			responseDump, err := httputil.DumpResponse(resp, true)
			if err != nil {
				fmt.Println(err)
			}
			httpLog.Println("\n---Response---\n" + string(responseDump))
		}

		//打印出header
		//for a,b := range(resp.Header) {
		//  fmt.Printf("%s : %s\n",a,b)
		//}

		var body []byte
		switch resp.Header.Get("Content-Encoding") {
		case "gzip":
			reader, err := gzip.NewReader(resp.Body)
			if err != nil {
				log.Panicln("[error]CacheHTTP read response:", err.Error())
			}
			for {
				buf := make([]byte, 1024)
				n, err := reader.Read(buf)

				if err != nil && err != io.EOF {
					panic(err)
				}

				if n == 0 {
					break
				}
				body = append(body, buf...)
			}
		default:
			body, err = io.ReadAll(resp.Body)
			if err != nil {
				log.Panicln("[error]CacheHTTP read response:", err.Error())
			}
		}

		httpRes.HTTPStatus = resp.Status
		httpRes.HTTPStatusCode = resp.StatusCode
		httpRes.ContentLength = resp.ContentLength
		httpRes.Body = body
		ret <- httpRes
		//close(ret)

		return nil

	}, func(err error) error {
		log.Println("[error]hystrix fallback")
		//放入redis重试池
		type RedisHTTPRetryPool struct {
			Request    HTTPRequest
			RetryCount int
		}

		_, errcache := c.Cache.LPUSH("http_retry_pool", RedisHTTPRetryPool{Request: rq})
		if errcache != nil {
			println(errcache.Error())
		}
		ret <- httpRes
		return nil
	})

	return nil
}

// Run 执行方法,支持http多协程请求并缓存
func (c *HTTPQuery) Run(cher HTTPQueryer) (map[string]interface{}, error) {
	type CacheData struct {
		NeedCache bool
		CacheKey  string
		Rtype     string
		HTTPError bool
		Cachedata []byte
	}

	UNIQID := c.GenUniqID()

	expire := cher.GetCacheExpire()
	ch := cher.GetBuilder()
	cher.InitHTTP()

	AllCacheData := make(map[string]CacheData) //key UniqID value:CacheData

	debug := cher.GetDebugInfo()

	debug.Add(c.Cache.Show())

	var NeedHTTPSum int
	for _, eachhttp := range ch.Requests {
		//添加letsgo User-agent
		if len(eachhttp.Header) == 0 {
			eachhttp.Header = make(map[string]string)
		}
		eachhttp.Header["User-Agent"] = "Letsgo-HTTP-Agent"

		cachekey := "HTTP_" + eachhttp.UniqID
		var retdata []byte
		if eachhttp.NeedCache {
			if isget, err := c.Cache.Get(cachekey, &retdata); isget != true { //cache miss or error
				if err != nil {
					return nil, fmt.Errorf("[error]CacheHTTP get cache:%s %s", err.Error(), UNIQID)
				}

				debug.Add(fmt.Sprintf("Cache Miss: %s", cachekey))
				//debug.Add(fmt.Sprintf("CacheHTTP UniqID: %s", uniqid))

				go c.SampleHTTPQuery(eachhttp, debug, ch.ResponseCH)
				c.AddCounter()
				AllCacheData[eachhttp.UniqID] = CacheData{NeedCache: eachhttp.NeedCache, CacheKey: cachekey, Rtype: eachhttp.Rtype, HTTPError: false, Cachedata: nil}
				NeedHTTPSum++
			} else { //get cache
				debug.Add(fmt.Sprintf("Cache Get: %s", cachekey))
				AllCacheData[eachhttp.UniqID] = CacheData{NeedCache: false, CacheKey: cachekey, Rtype: eachhttp.Rtype, HTTPError: false, Cachedata: retdata}
			}
		} else {
			//goroutine http
			go c.SampleHTTPQuery(eachhttp, debug, ch.ResponseCH)
			c.AddCounter()
			AllCacheData[eachhttp.UniqID] = CacheData{NeedCache: eachhttp.NeedCache, CacheKey: "", Rtype: eachhttp.Rtype, HTTPError: false, Cachedata: nil}
			NeedHTTPSum++
		}
	}

	for NeedHTTPSum > 0 {
		select {
		case i, ok := <-ch.ResponseCH:
			if ok {
				//fmt.Println("CacheHTTP channel receive data:", string(i.Body))
				data := AllCacheData[i.UniqID]
				data.Cachedata = i.Body
				if i.ResponseStatus == -1 {
					data.HTTPError = true
				}
				AllCacheData[i.UniqID] = data
			} else {
				return nil, fmt.Errorf("[error]CacheHTTP channel closed before reading: %s", UNIQID)
			}
		case <-time.After(config.CACHEHTTP_SELECT_TIMEOUT):
			return nil, fmt.Errorf("[error]CacheHTTP channel timeout after %d second: %s", config.CACHEHTTP_SELECT_TIMEOUT, UNIQID)
		}
		NeedHTTPSum--
		c.SubCounter()
	}

	retdata := make(map[string]interface{}) // key:UniqID value:interface{}
	for uniqid, cachedata := range AllCacheData {
		if cachedata.CacheKey != "" && cachedata.NeedCache == true {
			//downgrade get shorter TTL 60 second
			cacheexpire := expire
			if cachedata.HTTPError == true {
				cacheexpire = config.CACHEHTTP_DOWNGRADE_CACHE_EXPIRE
			}

			err := c.Cache.Set(cachedata.CacheKey, &cachedata.Cachedata, cacheexpire)
			if err != nil {
				return nil, fmt.Errorf("[error]CacheHTTP set cache:%s %s", err.Error(), UNIQID)
			}
			debug.Add(fmt.Sprintf("Cache Set: %s TTL: %d", cachedata.CacheKey, cacheexpire))
		}

		var ret interface{}
		var err error
		switch cachedata.Rtype {
		case "JSON":
			ret, err = simplejson.NewJson(cachedata.Cachedata)
			if err != nil {
				return nil, fmt.Errorf("[error]CacheHTTP DataToJson:%s %s", err.Error(), UNIQID)
			}
		case "HTML":
			ret = string(cachedata.Cachedata)
		default:
			ret = string(cachedata.Cachedata)
		}
		retdata[uniqid] = ret
	}

	return retdata, nil
}
