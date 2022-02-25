package letsgo

//MicroserviceClienter
type MicroserviceClienter interface {
	Init() error
	RegisterService(service_name string, service_port int) error
	DeregisterService(service_name string) error
	DeregisterAllService() error
	ServiceDiscovery(service_name string) (string, error)
	IsActive() bool
	DeregisterBeforeExit(todo bool) bool
	IsDeregisterBeforeExit() bool
}
