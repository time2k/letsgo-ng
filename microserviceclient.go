package letsgo

//MicroServiceClienter
type MicroServiceClienter interface {
	Init()
	RegisterService(service_name string, service_port int)
	DeregisterService(service_name string)
	ServiceFind(service_name string) string
}
