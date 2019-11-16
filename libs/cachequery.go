package Libs

import (
	"bytes"
	"database/sql"
	"encoding/gob"
	"fmt"
	"reflect"
	"strings"
	"sync"

	_ "github.com/go-sql-driver/mysql"
)

//CacheQueryer 接口描述
type CacheQueryer interface {
	GetCacheKey() string
	GetCacheExpire() int32
	GetSQL() *SQLmodel
	GetDebugInfo() *DebugModel
	GetDbname() string
}

//LCacheQuery 结构体
type LCacheQuery struct {
	DBset          DBC
	Cache          *LCache
	SQLcounter     int
	SQLcounterLock sync.Mutex
	RWflag         int
	RWflagLock     sync.Mutex
}

//NewCacheQuery 返回一个LCacheQuery结构体指针
func NewCacheQuery() *LCacheQuery {
	return &LCacheQuery{}
}

//SetDBset 设置db连接集
func (c *LCacheQuery) SetDBset(dbset DBC) {
	c.DBset = dbset
}

//SetCache 设置cache
func (c *LCacheQuery) SetCache(cache *LCache) {
	c.Cache = cache
}

//AddCounter 内置计数器++
func (c *LCacheQuery) AddCounter() {
	c.SQLcounterLock.Lock()
	defer c.SQLcounterLock.Unlock()
	c.SQLcounter++
}

//SubCounter 内置计数器--
func (c *LCacheQuery) SubCounter() {
	c.SQLcounterLock.Lock()
	defer c.SQLcounterLock.Unlock()
	c.SQLcounter--
}

//SelectOne 单条查询方法
func (c *LCacheQuery) SelectOne(cqer CacheQueryer) (bool, error) {
	c.AddCounter()
	defer c.SubCounter()

	sql := cqer.GetSQL()
	SQL := sql.SQL
	SQLcondition := sql.SQLcondition
	DataStruct := sql.DataStruct
	cache_key := cqer.GetCacheKey()
	cache_expire := cqer.GetCacheExpire()
	debug := cqer.GetDebugInfo()
	DbName := cqer.GetDbname()

	debug.Add(c.Cache.Show())

	if isget, err := c.Cache.Get(cache_key, DataStruct); isget != true { //cache miss or error
		if err != nil {
			return false, fmt.Errorf("[error]CacheQuery get cache: %s", err.Error())
		}

		debug.Add(fmt.Sprintf("Cache Miss: %s", cache_key))
		dbconn, err := c.ReadMSBalancer(DbName)
		if err != nil {
			return false, fmt.Errorf("[error]CacheQuery: %s", err.Error())
		}
		rows, err := dbconn.Query(SQL, SQLcondition...)
		if err != nil {
			return false, fmt.Errorf("[error]CacheQuery DB query action: %s", err.Error())
		}
		defer rows.Close()

		debug.Add(fmt.Sprintf("Get DB Query: %s , Query Condition: %s", SQL, SQLcondition))

		if rows.Next() {
			err = rows.Err()
			if err != nil {
				return false, fmt.Errorf("[error]CacheQuery DB rows.next action: %s", err.Error())
			}
			s := reflect.Indirect(reflect.ValueOf(DataStruct).Elem())
			len := s.NumField()
			scan_p := make([]interface{}, len)
			for k := 0; k < len; k++ {
				scan_p[k] = s.Field(k).Addr().Interface()
			}
			err = rows.Scan(scan_p...)

			if err != nil {
				return false, fmt.Errorf("[error]CacheQuery DB scan action: %s", err.Error())
			}

			err = c.Cache.Set(cache_key, DataStruct, cache_expire)
			if err != nil {
				return false, fmt.Errorf("[error]CacheQuery set cache: %s", err.Error())
			}

			debug.Add(fmt.Sprintf("Cache Set: %s TTL: %d", cache_key, cache_expire))
		} else {
			return false, nil
		}
	} else { //get cache
		debug.Add(fmt.Sprintf("Cache Get: %s", cache_key))
	}
	return true, nil
}

//SelectMulti 多条查询方法
func (c *LCacheQuery) SelectMulti(cqer CacheQueryer) (bool, error) {
	c.AddCounter()
	defer c.SubCounter()

	sql := cqer.GetSQL()
	SQL := sql.SQL
	SQLcondition := sql.SQLcondition
	DataStruct := sql.DataStruct
	cache_key := cqer.GetCacheKey()
	cache_expire := cqer.GetCacheExpire()
	debug := cqer.GetDebugInfo()
	DbName := cqer.GetDbname()

	debug.Add(c.Cache.Show())

	if isget, err := c.Cache.Get(cache_key, DataStruct); isget != true { //cache miss or error
		if err != nil {
			return false, fmt.Errorf("[CacheQuery]get cache: %s", err.Error())
		}

		debug.Add(fmt.Sprintf("Cache Miss: %s", cache_key))

		dbconn, err := c.ReadMSBalancer(DbName)
		if err != nil {
			return false, fmt.Errorf("[error]CacheQuery: %s", err.Error())
		}
		rows, err := dbconn.Query(SQL, SQLcondition...)

		if err != nil {
			return false, fmt.Errorf("[CacheQuery]DB query action: %s", err.Error())
		}
		defer rows.Close()

		debug.Add(fmt.Sprintf("Get DB Query: %s , Query Condition: %s", SQL, SQLcondition))

		if reflect.TypeOf(DataStruct).Kind() != reflect.Ptr && reflect.TypeOf(DataStruct).Elem().Kind() != reflect.Slice {
			return false, fmt.Errorf("[CacheQuery]Datastruct must be a Pointer of slice")
		}

		news := reflect.TypeOf(DataStruct).Elem().Elem()
		rowc := 0
		for rows.Next() {
			err = rows.Err()
			if err != nil {
				return false, fmt.Errorf("[CacheQuery]DB rows.next action: %s", err.Error())
			}

			dp := reflect.New(news)
			dx := reflect.Indirect(dp)

			len := dx.NumField()
			scan_p := make([]interface{}, len)
			for k := 0; k < len; k++ {
				scan_p[k] = dx.Field(k).Addr().Interface()
			}
			err = rows.Scan(scan_p...)
			if err != nil {
				return false, fmt.Errorf("[CacheQuery]DB scan action: %s", err.Error())
			}
			//reflect type of append into interface{} of a pointer of slice
			reflect.ValueOf(DataStruct).Elem().Set(reflect.Append(reflect.ValueOf(DataStruct).Elem(), dp.Elem()))

			rowc++
		}
		if rowc == 0 {
			return false, nil
		}
		err = c.Cache.Set(cache_key, DataStruct, cache_expire)
		if err != nil {
			return false, fmt.Errorf("[CacheQuery]set cache: %s", err.Error())
		}
		debug.Add(fmt.Sprintf("Cache Set: %s TTL: %d", cache_key, cache_expire))
	} else { //get cache
		debug.Add(fmt.Sprintf("Cache Get: %s", cache_key))
	}

	return true, nil
}

//EXEC 数据执行类 insert update 等请用此函数
func (c *LCacheQuery) EXEC(cqer CacheQueryer) (int64, error) {
	c.AddCounter()
	defer c.SubCounter()

	sql := cqer.GetSQL()
	SQL := sql.SQL
	SQLcondition := sql.SQLcondition
	debug := cqer.GetDebugInfo()

	debug.Add(fmt.Sprintf("EXEC DB Query: %s , Query Condition: %s", SQL, SQLcondition))
	stmt, err := c.DBset[cqer.GetDbname()].Master.Prepare(SQL)
	if err != nil {
		return 0, fmt.Errorf("[error]CacheQuery stmt sql: %s", err.Error())
	}
	defer stmt.Close()
	res, err2 := stmt.Exec(SQLcondition...)
	if err2 != nil {
		return 0, fmt.Errorf("[error]CacheQuery exe sql: %s", err2.Error())
	}
	if strings.Contains(SQL, "INSERT") || strings.Contains(SQL, "insert") {
		id, err := res.LastInsertId()
		if err != nil {
			return 0, err
		}
		return id, nil
	}

	if strings.Contains(SQL, "UPDATE") || strings.Contains(SQL, "update") || strings.Contains(SQL, "DELETE") || strings.Contains(SQL, "delete") {
		num, err := res.RowsAffected()
		if err != nil {
			return 0, err
		}
		return num, nil
	}

	return 0, nil
}

//GetTX 事务类，返回一个tx连接
func (c *LCacheQuery) GetTX(cqer CacheQueryer) (*sql.Tx, error) {
	return c.DBset[cqer.GetDbname()].Master.Begin()
}

//ReadMSBalancer 轮询使用master或者slave进行查询
func (c *LCacheQuery) ReadMSBalancer(DbName string) (*sql.DB, error) {
	c.RWflagLock.Lock()
	defer c.RWflagLock.Unlock()
	if _, ok := c.DBset[DbName]; !ok { //key不存在
		return nil, fmt.Errorf("[error]CacheQuery ReadMSBalancer: can't find this db config '%s'", DbName)
	}
	if c.RWflag == 0 {
		c.RWflag = 1
		return c.DBset[DbName].Master, nil
	}

	c.RWflag = 0
	if c.DBset[DbName].Slave != nil {
		return c.DBset[DbName].Slave, nil
	}

	return c.DBset[DbName].Master, nil
}

//deepCopy 深拷贝方法
func deepCopy(dst, src interface{}) error {
	var buf bytes.Buffer
	if err := gob.NewEncoder(&buf).Encode(src); err != nil {
		return err
	}
	return gob.NewDecoder(bytes.NewBuffer(buf.Bytes())).Decode(dst)
}
