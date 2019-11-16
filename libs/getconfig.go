package Libs

import (
	"Letsgo2/Lconfig"
	"Letsgo2/Ltypedef"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"
)

//SingleConfigHashMap 单配置的hashmap结构
type SingleConfigHashMap struct {
	RwLock    sync.RWMutex
	Config    map[string]string
	ConfigTTL int64
}

//ConfigSet 多个配置的hashmap结构
type ConfigSet struct {
	RwLock    sync.RWMutex
	ConfigSet map[string]*SingleConfigHashMap
}

//CONFIG_SET 常驻内存的变量
var CONFIG_SET = new(ConfigSet)

//InitConfig 初始化config
func InitConfig() {
	CONFIG_SET.ConfigSet = make(map[string]*SingleConfigHashMap)
}

//Get SingleConfig 获取具体配置项方法
func (SC *SingleConfigHashMap) Get(ConfigName string) (string, bool) {
	SC.RwLock.RLock()
	defer SC.RwLock.RUnlock()
	if val, ok := SC.Config[ConfigName]; ok {
		return val, true
	}
	return "", false
}

//Set SingleConfig 设置方法
func (SC *SingleConfigHashMap) Set(ConfigName string, ConfigValue string) {
	SC.RwLock.Lock()
	defer SC.RwLock.Unlock()
	SC.Config[ConfigName] = ConfigValue
}

//GetConfig 获取配置集方法
func GetConfig(letsgo *Letsgo, ConfigSetName string) *SingleConfigHashMap {
	CONFIG_SET.RwLock.Lock() //因为包含写操作，要加写锁
	defer CONFIG_SET.RwLock.Unlock()

	if _, ok := CONFIG_SET.ConfigSet[ConfigSetName]; !ok { //没有从set获取到config_name对应的config
		CONFIG_SET.ConfigSet[ConfigSetName] = &SingleConfigHashMap{Config: make(map[string]string)}
	}

	if len(CONFIG_SET.ConfigSet[ConfigSetName].Config) == 0 { //获取到的config长度为0
		err := SetConfig(letsgo, ConfigSetName)
		if err != nil {
			letsgo.Logger.Println("Lib getconfig error:", err.Error())
		}
	} else if CONFIG_SET.ConfigSet[ConfigSetName].ConfigTTL+600 <= time.Now().Unix() { //过期了
		err := SetConfig(letsgo, ConfigSetName)
		if err != nil {
			letsgo.Logger.Println("Lib getconfig error:", err.Error())
		}
	}

	return CONFIG_SET.ConfigSet[ConfigSetName]
}

//SetConfig 设置配置集方法
func SetConfig(letsgo *Letsgo, ConfigSetName string) error {
	switch ConfigSetName {
	case "EXAMPLE_SET":
		//init
		data := new(Ltypedef.ExampleConfigSet)
		var debuginfo []DebugModel
		//组合成主键
		cache_key := "example_set"

		model := NewLModel()
		model.SetCacheKey(cache_key)
		model.SetCacheExpire(Lconfig.CONFIG_CACHE_EXPIRE)
		model.SetSQL("SELECT id,value FROM example_confiset ORDER BY id ASC")
		model.SetDataStruct(&data.Set)
		model.SetDbname("vrs")

		data_exists, err := letsgo.CacheQuery.SelectMulti(model)
		if err != nil {
			return fmt.Errorf("Lib getconfig get db error: %s", err.Error())
		}

		debuginfo = append(debuginfo, model.DebugModel)

		if data_exists == false {
			return fmt.Errorf("Lib getconfig get db data return no data")
		}

		CONFIG_SET.ConfigSet[ConfigSetName].RwLock.Lock()
		for _, v := range data.Set {
			CONFIG_SET.ConfigSet[ConfigSetName].Config[v.ID] = v.Value
		}
		CONFIG_SET.ConfigSet[ConfigSetName].RwLock.Unlock()

		CONFIG_SET.ConfigSet[ConfigSetName].ConfigTTL = time.Now().Unix()
		letsgo.Logger.Println("Lib getconfig load", ConfigSetName, "success!")
		return nil

	default:
		return fmt.Errorf("Lib getconfig get no registed config")
	}

}

var (
	once sync.Once
	conf *Conf
)

type Conf struct {
	Common
	Mysql       Mysql
	Memcached   Memcached
	CacheHttp   CacheHttp
	Schedule    Schedule
	Redis       Redis
	AreaListRev map[string]string
	Demote      map[string]string
}

type Common struct {
	SERVER_PORT               string
	SERVER_READTIMEOUT        string
	SERVER_WRITETIMEOUT       string
	VERSION                   string
	ACCESS_LOG                string
	ERROR_LOG                 string
	BASEAUTH                  string
	BASEAUTH_SALTKEY          string
	EXAMPLE_INFO_CACHE_EXPIRE string
	CONFIG_CACHE_EXPIRE       string
	StatusOk                  string
	StatusNoData              string
	StatusParamsNoValid       string
	StatusError               string
}

type Mysql struct {
	Vrs MSlave
}

type MSlave struct {
	Master *DB
	Slave  *[]DB
}

type DB struct {
	Host              string
	Port              string
	User              string
	Password          string
	DBname            string
	DBcharset         string
	DBconnMaxConns    string
	DBconnMaxIdles    string
	DBconnMaxLifeTime string
}

type Memcached struct {
	Expire    int32
	Timeout   int
	IdleConns int
	Servers   *[]MServer
}

type MServer struct {
	Host string
	Port string
}

type CacheHttp struct {
	DialTimeout          int
	ResponseTimeout      int
	ChannelBufferLen     int
	DowngradeCacheExpire int32
	SelectTimeout        int
}

type Schedule struct {
	ChannelBufferLen int
	MaxConcurrent    int
	SelectTimeout    int
}

type Redis struct {
	Servers *[]RServer
}

type RServer struct {
	Host           string
	Port           string
	Password       string
	ConnectTimeout int
	ReadTimeout    int
}

type AreaListRev struct {
	Access_Log string
	Error_Log  string
}

var IsChange bool = false

func GetViperConfig() (*Conf, error) {
	once.Do(func() {
		loadConfig(false)
	})

	viper.WatchConfig()
	viper.OnConfigChange(func(e fsnotify.Event) {
		loadConfig(true)
	})

	return conf, nil
}

func loadConfig(is_change bool) {
	IsChange = is_change

	// 当前程序的根目录，配置文件应该按照实际目录进行修改
	root := GetCurrentDirectory()
	root = "."

	viper.SetConfigName("config")
	viper.AddConfigPath(root + "/src/Letsgo2/Lconfig")

	if err := viper.ReadInConfig(); err != nil {
		conf = nil
	}

	viper.SetConfigType("toml")

	if err := viper.Unmarshal(&conf); err != nil {
		conf = nil
	}
}

func GetCurrentDirectory() string {
	dir, err := filepath.Abs(filepath.Dir(os.Args[0])) //返回绝对路径  filepath.Dir(os.Args[0])去除最后一个元素的路径
	if err != nil {
		return "."
	}

	last := strings.LastIndex(dir, "/")
	return dir[0:last]
}
