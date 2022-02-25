package letsgo

import (
	"fmt"
	"log"
	"net/rpc"
	"net/rpc/jsonrpc"
	"sync"
)

//RPCconfig rpc服务器配置结构体
type RPCconfig struct {
	Network          string //网络 可以是tcp udp或者http
	Address          string //具体地址及端口
	MicroserviceName string //rpc的微服务名称
}

//JSONRPCClient 结构体
type JSONRPCClient struct {
	Service map[string]RPCconfig
	Lock    sync.Mutex
}

//NewJSONRPCClient 返回一个JSONRPCClient结构体指针
func NewJSONRPCClient() *JSONRPCClient {
	return &JSONRPCClient{}
}

//Init JSONRPCClient初始化
func (c *JSONRPCClient) Init() {
	c.Service = make(map[string]RPCconfig)
}

//Set 设置config
func (c *JSONRPCClient) Set(service, network, address, microservice_name string) error {
	c.Lock.Lock()
	c.Service[service] = RPCconfig{Network: network, Address: address, MicroserviceName: microservice_name}
	c.Lock.Unlock()
	return nil
}

//Dial 连接到一个rpc服务器
func (c *JSONRPCClient) Dial(service string) (*rpc.Client, error) {
	thisservice, ok := c.Service[service]
	if !ok {
		return nil, fmt.Errorf("[error]jsonrpc Call unknown service: %s", service)
	}

	addr := "" //ip:port
	//如果配置了微服务，优先使用服务发现
	if Default.MicroserviceClient != nil && Default.MicroserviceClient.IsActive() {
		var err error
		addr, err = Default.MicroserviceClient.ServiceDiscovery(thisservice.MicroserviceName)
		if err != nil {
			return nil, fmt.Errorf("[error]jsonrpc ServiceDiscovery error:", err.Error())
		}
		log.Println("dial use microservice discovery, addr:", addr)
	} else {
		addr = thisservice.Address
		log.Println("dial use config address, addr:", addr)
	}

	client, err := jsonrpc.Dial(thisservice.Network, addr)
	if err != nil {
		return nil, fmt.Errorf("[error]jsonrpc dial: %s", err.Error())
	}

	return client, nil
}

//DialWithMicroserviceFind 使用微服务发现并连接到rpc服务器
func (c *JSONRPCClient) DialWithServiceDiscovery(service string) (*rpc.Client, error) {
	thisservice, ok := c.Service[service]
	if !ok {
		return nil, fmt.Errorf("[error]jsonrpc Call unknown service: %s", service)
	}

	if Default.MicroserviceClient == nil {
		return nil, fmt.Errorf("[error]jsonrpc need init MicroserviceClient")
	}

	if !Default.MicroserviceClient.IsActive() {
		return nil, fmt.Errorf("[error]jsonrpc need active MicroserviceClient")
	}

	addr, err := Default.MicroserviceClient.ServiceDiscovery(thisservice.MicroserviceName)
	if err != nil {
		return nil, fmt.Errorf("[error]jsonrpc ServiceDiscovery error:", err.Error())
	}

	client, err := jsonrpc.Dial(thisservice.Network, addr)
	if err != nil {
		return nil, fmt.Errorf("[error]jsonrpc dial: %s", err.Error())
	}

	return client, nil
}
