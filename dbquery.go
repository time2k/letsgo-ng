package letsgo

import (
	"bytes"
	"database/sql"
	"encoding/gob"
	"fmt"
	"reflect"
	"strings"
	"sync"

	_ "github.com/go-sql-driver/mysql" //mysql
)

// DBQueryer 接口描述
type DBQueryer interface {
	IsUseCache() bool
	GetCacheKey() string
	GetCacheExpire() int32
	GetBuilder() *DBQueryBuilder
	GetDebugInfo() *DebugInfo
	GetDbname() string
}

// DBSet 支持1主1从的DBset
type DBSet struct {
	Master *sql.DB
	Slave  *sql.DB
}

// DBC DBSet集合
type DBC map[string]DBSet

// DBQuery 结构体
type DBQuery struct {
	DBset          DBC
	Cache          *Cache
	SQLcounter     int
	SQLcounterLock sync.Mutex
	RWflag         int
	RWflagLock     sync.Mutex
}

// newDBQuery 返回一个DBQuery结构体指针
func newDBQuery() *DBQuery {
	return &DBQuery{}
}

// SetDBset 设置db连接集
func (c *DBQuery) SetDBset(dbset DBC) {
	c.DBset = dbset
}

// SetCache 设置cache
func (c *DBQuery) SetCache(cache *Cache) {
	c.Cache = cache
}

// AddCounter 内置计数器++
func (c *DBQuery) AddCounter() {
	c.SQLcounterLock.Lock()
	defer c.SQLcounterLock.Unlock()
	c.SQLcounter++
}

// SubCounter 内置计数器--
func (c *DBQuery) SubCounter() {
	c.SQLcounterLock.Lock()
	defer c.SQLcounterLock.Unlock()
	c.SQLcounter--
}

// SelectOne 单条查询方法
func (c *DBQuery) SelectOne(cqer DBQueryer) (bool, error) {
	c.AddCounter()
	defer c.SubCounter()

	DB := cqer.GetBuilder()
	SQL := DB.SQL
	SQLcondition := DB.SQLcondition
	Result := DB.Result
	CacheKey := cqer.GetCacheKey()
	CacheExpire := cqer.GetCacheExpire()
	debug := cqer.GetDebugInfo()
	DbName := cqer.GetDbname()
	UseCache := cqer.IsUseCache()

	//Result Must Be A ptr to slice
	if reflect.TypeOf(Result).Kind() != reflect.Ptr {
		return false, fmt.Errorf("[CacheQuery]Result must be a Pointer")
	}
	//result type element and result value element
	rtype := reflect.TypeOf(Result).Elem()
	rvalue := reflect.ValueOf(Result).Elem()

	if UseCache == true { //do use cache
		if isget, err := c.Cache.Get(CacheKey, Result); isget != true { //cache miss or error
			if err != nil {
				return false, fmt.Errorf("[error]CacheQuery get cache: %s", err.Error())
			}
			debug.Add(fmt.Sprintf("Cache Miss: %s", CacheKey))
		} else {
			debug.Add(fmt.Sprintf("Cache Get: %s", CacheKey))

			return true, nil
		}
	}

	dbconn, err := c.ReadMSBalancer(DbName)
	if err != nil {
		return false, fmt.Errorf("[error]CacheQuery: %s", err.Error())
	}
	rows, err := dbconn.Query(SQL, SQLcondition...)
	if err != nil {
		return false, fmt.Errorf("[error]CacheQuery DB query action: %w", err)
	}
	defer rows.Close()

	debug.Add(fmt.Sprintf("Get DB Query: %s , Query Condition: %v", SQL, SQLcondition))

	if rows.Next() {
		err := rows.Err()
		if err != nil {
			return false, fmt.Errorf("[error]CacheQuery DB rows.next action: %w", err)
		}
		if rtype.Kind() == reflect.Struct {
			//s := reflect.Indirect(rve)
			len := rvalue.NumField()
			scanp := make([]interface{}, len)
			for k := 0; k < len; k++ {
				scanp[k] = rvalue.Field(k).Addr().Interface()
			}
			err = rows.Scan(scanp...)
			if err != nil {
				return false, fmt.Errorf("[error]CacheQuery DB scan action: %w", err)
			}
		} else {
			err = rows.Scan(rvalue.Addr().Interface())
			if err != nil {
				return false, fmt.Errorf("[error]CacheQuery DB scan action: %w", err)
			}
		}
		if UseCache == true { //do use cache
			err = c.Cache.Set(CacheKey, Result, CacheExpire)
			if err != nil {
				return false, fmt.Errorf("[error]CacheQuery set cache: %s", err.Error())
			}
			debug.Add(fmt.Sprintf("Cache Set: %s TTL: %d", CacheKey, CacheExpire))
		}
	} else {
		return false, nil
	}

	return true, nil
}

// SelectMulti 多条查询方法
func (c *DBQuery) SelectMulti(cqer DBQueryer) (bool, error) {
	c.AddCounter()
	defer c.SubCounter()

	DB := cqer.GetBuilder()
	SQL := DB.SQL
	SQLcondition := DB.SQLcondition
	Result := DB.Result
	CacheKey := cqer.GetCacheKey()
	CacheExpire := cqer.GetCacheExpire()
	debug := cqer.GetDebugInfo()
	DbName := cqer.GetDbname()
	UseCache := cqer.IsUseCache()

	//Result Must Be A ptr to slice
	if reflect.TypeOf(Result).Kind() != reflect.Ptr && reflect.TypeOf(Result).Elem().Kind() != reflect.Slice {
		return false, fmt.Errorf("[CacheQuery]Result must be a Pointer of slice")
	}
	//result type element and result value element
	rtype := reflect.TypeOf(Result).Elem().Elem()
	rvalue := reflect.ValueOf(Result).Elem() //indeed a slice

	if UseCache == true { //do use cache
		if isget, err := c.Cache.Get(CacheKey, Result); isget != true { //cache miss or error
			if err != nil {
				return false, fmt.Errorf("[CacheQuery]get cache: %s", err.Error())
			}

			debug.Add(fmt.Sprintf("Cache Miss: %s", CacheKey))
		} else {
			debug.Add(fmt.Sprintf("Cache Get: %s", CacheKey))
			return true, nil
		}
	}

	dbconn, err := c.ReadMSBalancer(DbName)
	if err != nil {
		return false, fmt.Errorf("[error]CacheQuery: %s", err.Error())
	}
	rows, err := dbconn.Query(SQL, SQLcondition...)

	if err != nil {
		return false, fmt.Errorf("[CacheQuery]DB query action: %w", err)
	}
	defer rows.Close()

	debug.Add(fmt.Sprintf("Get DB Query: %s , Query Condition: %v", SQL, SQLcondition))

	rowc := 0

	for rows.Next() {
		err := rows.Err()
		if err != nil {
			return false, fmt.Errorf("[CacheQuery]DB rows.next action: %w", err)
		}

		dp := reflect.New(rtype)
		dx := reflect.Indirect(dp)

		if rtype.Kind() == reflect.Struct { //like "select a,b,c from" and result like []struct {a int,b string,c float}
			len := dx.NumField()
			scanp := make([]interface{}, len)
			for k := 0; k < len; k++ {
				scanp[k] = dx.Field(k).Addr().Interface()
			}
			err := rows.Scan(scanp...)
			if err != nil {
				return false, fmt.Errorf("[CacheQuery]DB scan action: %w", err)
			}
		} else {
			err := rows.Scan(dx.Addr().Interface())
			if err != nil {
				return false, fmt.Errorf("[CacheQuery]DB scan action: %w", err)
			}
		}
		//reflect type of append into interface{} of a pointer of slice
		rvalue.Set(reflect.Append(rvalue, dp.Elem()))

		rowc++
	}

	if rowc == 0 {
		return false, nil
	}
	if UseCache == true { //do use cache
		err = c.Cache.Set(CacheKey, Result, CacheExpire)
		if err != nil {
			return false, fmt.Errorf("[CacheQuery]set cache: %s", err.Error())
		}
		debug.Add(fmt.Sprintf("Cache Set: %s TTL: %d", CacheKey, CacheExpire))
	}

	return true, nil
}

// EXEC 数据执行类 insert update 等请用此函数
func (c *DBQuery) EXEC(cqer DBQueryer) (int64, error) {
	c.AddCounter()
	defer c.SubCounter()

	DB := cqer.GetBuilder()
	SQL := DB.SQL
	SQLcondition := DB.SQLcondition
	debug := cqer.GetDebugInfo()

	debug.Add(fmt.Sprintf("EXEC DB Query: %s , Query Condition: %s", SQL, SQLcondition))

	DbName := cqer.GetDbname()
	if _, ok := c.DBset[DbName]; !ok { //key不存在
		return 0, fmt.Errorf("[error]CacheQuery exec: can't find this db config '%s'", DbName)
	}

	stmt, err := c.DBset[DbName].Master.Prepare(SQL)
	if err != nil {
		return 0, fmt.Errorf("[error]CacheQuery stmt sql: %w", err)
	}
	defer stmt.Close()
	res, err2 := stmt.Exec(SQLcondition...)
	if err2 != nil {
		return 0, fmt.Errorf("[error]CacheQuery exe sql: %w", err2)
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

// GetTX 事务类，返回一个tx连接
func (c *DBQuery) GetTX(cqer DBQueryer) (*sql.Tx, error) {
	return c.DBset[cqer.GetDbname()].Master.Begin()
}

// ReadMSBalancer 轮询使用master或者slave进行查询
func (c *DBQuery) ReadMSBalancer(DbName string) (*sql.DB, error) {
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

// deepCopy 深拷贝方法
func deepCopy(dst, src interface{}) error {
	var buf bytes.Buffer
	if err := gob.NewEncoder(&buf).Encode(src); err != nil {
		return err
	}
	return gob.NewDecoder(bytes.NewBuffer(buf.Bytes())).Decode(dst)
}
