package letsgo

//MicroServiceClienter
type MicroServiceClienter interface {
	Init()
	RegisterService(service_name string, service_port int)
	DeregisterService(service_name string)
	DeregisterAllService()
	ServiceFind(service_name string) string
	IsActive() bool
}
