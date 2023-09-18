package naming

import "fmt"

type ServiceRegistration interface {
	ServiceID() string
	ServiceName() string
	// PublicAddress ip or domain
	PublicAddress() string
	DialURL() string
	GetProtocol() string
	GetNamespace() string
	GetTags() []string
	GetMeta() map[string]string
	String() string
}

// "ID": "qa-dfirst-zfirst-tgateway-172.16.235.145-0-8000",
// "Service": "tgateway",
// "Tags": [
// "ZONE:qa-dfirst-zfirst",
// "TMC_REGION:SH",
// "TMC_DOMAIN:g002-qa.tutormeetplus.com"
// ],
// "Address": "172.16.235.145",
// "Port": 8000,

type DefaultService struct {
	Id        string
	Name      string
	Address   string
	Port      int
	Protocol  string
	Namespace string
	Tags      []string
	Meta      map[string]string
}

func NewEntry(id, name, protocol string, address string, port int) ServiceRegistration {
	return &DefaultService{
		Id:       id,
		Name:     name,
		Address:  address,
		Port:     port,
		Protocol: protocol,
	}
}

func (e *DefaultService) String() string {
	return fmt.Sprintf("Id:%s,Name:%s,Address:%s,Port:%d,Ns:%s,Tags:%v,Meta:%v", e.Id, e.Name, e.Address, e.Port, e.Namespace, e.Tags, e.Meta)
}

func (e *DefaultService) GetMeta() map[string]string { return e.Meta }

func (e *DefaultService) GetTags() []string { return e.Tags }

func (e *DefaultService) DialURL() string {
	if e.Protocol == "tcp" {
		return fmt.Sprintf("%s:%d", e.Address, e.Port)
	}
	return fmt.Sprintf("%s://%s:%d", e.Protocol, e.Address, e.Port)
}

func (e *DefaultService) GetProtocol() string { return e.Protocol }

func (e *DefaultService) PublicAddress() string { return e.Address }

func (e *DefaultService) PublicPort() int { return e.Port }

func (e *DefaultService) ServiceName() string { return e.Name }

func (e *DefaultService) GetNamespace() string { return e.Namespace }

func (e *DefaultService) ServiceID() string {
	return e.Id
}
