package naming

import (
	"errors"

	"github.com/sjmshsh/HopeIM"
)

// errors
var (
	ErrNotFound = errors.New("service no found")
)

// Naming defined methods of the naming service
type Naming interface {
	Find(serviceName string, tags ...string) ([]HopeIM.ServiceRegistration, error)
	Subscribe(serviceName string, callback func(services []HopeIM.ServiceRegistration)) error
	Unsubscribe(serviceName string) error
	Register(service HopeIM.ServiceRegistration) error
	Deregister(serviceID string) error
}
