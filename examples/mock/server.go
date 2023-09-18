package mock

import (
	"github.com/sjmshsh/HopeIM"
	"github.com/sjmshsh/HopeIM/naming"
	"net/http"
)

type ServerDemo struct{}

func (s *ServerDemo) Start(id, protocol, addr string) {
	go func() {
		http.ListenAndServe("0.0.0.0:6060", nil)
	}()

	var srv HopeIM.Server
	service := &naming.DefaultService{
		Id:       id,
		Protocol: protocol,
	}
	if protocol == "ws" {

	} else if protocol == "tcp" {

	}
}
