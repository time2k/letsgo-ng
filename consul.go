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
	Lock      sync.Mutex
}

//NewConsulClient 返回一个ConsulClient结构体指针
func NewConsulClient() *ConsulClient {
	return &ConsulClient{}
}

//Init ConsulClient
func (c *ConsulClient) Init() {
	config := consulapi.DefaultConfig()

	client, err := consulapi.NewClient(config)
	if err != nil {
		log.Println("ConsulClient init error : ", err.Error())
	}

	c.Client = client
	c.Active = true
}

//使用网卡设备名interface获取ip
func GetInterfaceIP(name string) string {
	iface, err := net.InterfaceByName(name) //here your interface
	if err != nil {
		return ""
	}

	addrs, err := iface.Addrs()
	if err != nil {
		return ""
	}
	for _, address := range addrs {
		// check the address type and if it is not a loopback the display it
		if ipnet, ok := address.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				return ipnet.IP.String()
			}
		}
	}
	return ""
}

//注册服务
func (c *ConsulClient) RegisterService(service_name string, service_port int) {
	hostname, err := os.Hostname() //使用hostname作为serviceid
	if err != nil {
		log.Println("ConsulClient RegisterService get os.Hostname error:", err.Error())
	}

	internalip := GetInterfaceIP("eth0")
	if internalip == "" {
		internalip = GetInterfaceIP("en0")
		if internalip == "" {
			internalip = "127.0.0.1"
		}
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
		log.Println("ConsulClient RegisterService register server error : ", err.Error())
		return
	}
	c.Lock.Lock()
	c.ServiceID = append(c.ServiceID, registration.ID)
	c.Lock.Unlock()
}

//删除服务
func (c *ConsulClient) DeregisterService(service_name string) {
	hostname, err := os.Hostname()
	if err != nil {
		log.Println("ConsulClient DeregisterService get os.Hostname error:", err.Error())
	}

	err = c.Client.Agent().ServiceDeregister(service_name + ":" + hostname)
	if err != nil {
		log.Println("ConsulClient DeregisterService deregister service error : ", err.Error())
	}
}

//删除所有服务
func (c *ConsulClient) DeregisterAllService() {
	for _, id := range c.ServiceID {
		println(id)
		err := c.Client.Agent().ServiceDeregister(id)
		if err != nil {
			log.Println("ConsulClient DeregisterAllService deregister service error : ", err.Error())
		}
	}
}

//发现服务 返回ip:port字符串
func (c *ConsulClient) ServiceFind(service_name string) string {
	services, err := c.Client.Agent().Services()
	if err != nil {
		log.Println("ConsulClient ServiceFind get service error : ", err.Error())
	}
	if _, found := services[service_name]; !found {
		log.Println("ConsulClient ServiceFind service_name not found")
	}
	return fmt.Sprint(services[service_name].Address, ":", services[service_name].Port)
}

//微服务客户端是否活跃
func (c *ConsulClient) IsActive() bool {
	return c.Active
}
