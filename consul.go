package letsgo

import (
	"fmt"
	"log"
	"net"
	"os"
	"sync"

	consulapi "github.com/hashicorp/consul/api"
)

//ConsulClient 结构体
type ConsulClient struct {
	Client    *consulapi.Client
	Active    bool
	ServiceID []string
	DExit     bool
	Lock      sync.Mutex
}

//NewConsulClient 返回一个ConsulClient结构体指针
func NewConsulClient() *ConsulClient {
	return &ConsulClient{}
}

//Init ConsulClient
func (c *ConsulClient) Init() error {
	config := consulapi.DefaultConfig()

	client, err := consulapi.NewClient(config)
	if err != nil {
		log.Println("[error]ConsulClient init error : ", err.Error())
		return err
	}

	c.Client = client
	c.Active = true
	return nil
}

//使用网卡设备名interface获取ip
func GetInterfaceIP(name string) (string, error) {
	iface, err := net.InterfaceByName(name) //here your interface
	if err != nil {
		return "", err
	}

	addrs, err := iface.Addrs()
	if err != nil {
		return "", err
	}
	for _, address := range addrs {
		// check the address type and if it is not a loopback the display it
		if ipnet, ok := address.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				return ipnet.IP.String(), nil
			}
		}
	}
	return "", err
}

//使用网卡设备名interface获取ip,依次获取
func SeqGetInterfaceIP(name ...string) string {
	internalip := ""
	var err error
	for _, eachname := range name {
		internalip, err = GetInterfaceIP(eachname)
		if internalip != "" && err == nil {
			return internalip
		}
	}
	return ""
}

//注册服务
func (c *ConsulClient) RegisterService(service_name string, service_port int) error {
	hostname, err := os.Hostname() //使用hostname作为serviceid
	if err != nil {
		log.Println("[error]ConsulClient RegisterService get os.Hostname error:", err.Error())
		return err
	}

	internalip := SeqGetInterfaceIP("eth0", "en0")
	if internalip == "" {
		internalip = "127.0.0.1"
	}

	registration := new(consulapi.AgentServiceRegistration)
	registration.ID = service_name + ":" + hostname
	registration.Name = service_name
	registration.Port = service_port
	registration.Tags = []string{service_name + ":" + hostname}
	registration.Address = internalip
	registration.Check = &consulapi.AgentServiceCheck{
		TCP:                            fmt.Sprintf("%s:%d", registration.Address, service_port),
		Timeout:                        "3s",
		Interval:                       "5s",
		DeregisterCriticalServiceAfter: "30s", //check失败后30秒删除本服务
	}

	err = c.Client.Agent().ServiceRegister(registration)
	if err != nil {
		log.Println("[error]ConsulClient RegisterService register server error : ", err.Error())
		return err
	}
	c.Lock.Lock()
	c.ServiceID = append(c.ServiceID, registration.ID)
	c.Lock.Unlock()
	return nil
}

//删除服务
func (c *ConsulClient) DeregisterService(service_name string) error {
	hostname, err := os.Hostname()
	if err != nil {
		log.Println("[error]ConsulClient DeregisterService get os.Hostname error:", err.Error())
		return err
	}

	err = c.Client.Agent().ServiceDeregister(service_name + ":" + hostname)
	if err != nil {
		log.Println("[error]ConsulClient DeregisterService deregister service error : ", err.Error())
		return err
	}
	return nil
}

//删除所有服务
func (c *ConsulClient) DeregisterAllService() error {
	for _, id := range c.ServiceID {
		err := c.Client.Agent().ServiceDeregister(id)
		if err != nil {
			log.Println("[error]ConsulClient DeregisterAllService deregister service error : ", err.Error())
			return err
		}
	}
	return nil
}

//发现服务 返回ip:port字符串 建议还是使用dns方式
func (c *ConsulClient) ServiceDiscovery(service_name string) (string, error) {
	services, _, err := c.Client.Health().Service(service_name, "", true, nil)
	if err != nil {
		log.Println("[error]ConsulClient ServiceDiscovery get service error : ", err.Error())
		return "", err
	}

	//随机使用一个service
	index := RandNum(len(services))

	return fmt.Sprint(services[index].Service.Address, ":", services[index].Service.Port), nil
}

//微服务客户端是否活跃
func (c *ConsulClient) IsActive() bool {
	return c.Active
}

//设置微服务是否server退出前删除
func (c *ConsulClient) DeregisterBeforeExit(todo bool) bool {
	c.DExit = todo
	return todo
}

//微服务是否server退出前删除
func (c *ConsulClient) IsDeregisterBeforeExit() bool {
	return c.DExit
}
