package letsgo

import (
	"fmt"
	"net/rpc"
	"net/rpc/jsonrpc"
	"sync"
)

//RPCconfig rpc服务器配置结构体
type RPCconfig struct {
	Network string //网络 可以是tcp udp或者http
	Address string //具体地址及端口
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
func (c *JSONRPCClient) Set(service, network, address string) error {
	c.Lock.Lock()
	c.Service[service] = RPCconfig{Network: network, Address: address}
	c.Lock.Unlock()
	return nil
}

//Dial 连接到一个rpc服务器
func (c *JSONRPCClient) Dial(service string) (*rpc.Client, error) {
	thisservice, ok := c.Service[service]
	if ok != true {
		return nil, fmt.Errorf("[error]jsonrpc Call unknown service: %s", service)
	}

	client, err := jsonrpc.Dial(thisservice.Network, thisservice.Address)
	if err != nil {
		return nil, fmt.Errorf("[error]jsonrpc dial: %s", err.Error())
	}

	return client, nil
}
