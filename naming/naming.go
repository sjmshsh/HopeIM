package naming

import "errors"

var (
	ErrNotFound = errors.New("service no found")
)

type Naming interface {
	Find(serviceName string) ([]ServiceRegistration, error)
	Remove(serviceName, serviceID string) error
	Register(registration ServiceRegistration) error
	Deregister(serviceID string) error
}
