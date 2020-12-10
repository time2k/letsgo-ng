package letsgo

import (
	"fmt"
	"sync"
	"time"
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

		CONFIG_SET.ConfigSet[ConfigSetName].RwLock.Lock()
		CONFIG_SET.ConfigSet[ConfigSetName].Config["ex"] = "123"
		CONFIG_SET.ConfigSet[ConfigSetName].RwLock.Unlock()

		CONFIG_SET.ConfigSet[ConfigSetName].ConfigTTL = time.Now().Unix()
		letsgo.Logger.Println("Lib getconfig load", ConfigSetName, "success!")
		return nil

	default:
		return fmt.Errorf("Lib getconfig get no registed config")
	}

}
