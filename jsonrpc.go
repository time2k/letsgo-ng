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

//LJsonrpcClient 结构体
type LJsonrpcClient struct {
	Service map[string]RPCconfig
	Lock    sync.Mutex
}

//NewLJsonrpcClient 返回一个LJsonrpcClient结构体指针
func NewLJsonrpcClient() *LJsonrpcClient {
	return &LJsonrpcClient{}
}

//Init LJsonrpcClient初始化
func (c *LJsonrpcClient) Init() {
	c.Service = make(map[string]RPCconfig)
}

func (rpc *LJsonrpcClient) Set(service, network, address string) error {
	rpc.Lock.Lock()
	rpc.Service[service] = RPCconfig{Network: network, Address: address}
	rpc.Lock.Unlock()
	return nil
}

//Dial 连接到一个rpc服务器
func (rpc *LJsonrpcClient) Dial(service string) (*rpc.Client, error) {
	this_service, ok := rpc.Service[service]
	if ok != true {
		return nil, fmt.Errorf("[error]jsonrpc Call unknown service: %s", service)
	}
	client, err := jsonrpc.Dial(this_service.Network, this_service.Address)
	if err != nil {
		return nil, fmt.Errorf("[error]jsonrpc dial: %s", err.Error())
	}

	return client, nil
}
