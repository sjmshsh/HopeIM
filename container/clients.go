package container

import (
	"github.com/sjmshsh/HopeIM"
	"github.com/sjmshsh/HopeIM/logger"
	"sync"
)

type ClientMap interface {
	Add(client HopeIM.Client)
	Remove(id string)
	Get(id string) (client HopeIM.Client, ok bool)
	Services(kvs ...string) []HopeIM.Service
}

type ClientsImpl struct {
	clients *sync.Map
}

func NewClients(num int) ClientMap {
	return &ClientsImpl{
		clients: new(sync.Map),
	}
}

func (ch *ClientsImpl) Add(client HopeIM.Client) {
	if client.ServiceID() == "" {
		logger.WithFields(logger.Fields{
			"module": "ClientsImpl",
		}).Error("client id is required")
	}
	ch.clients.Store(client.ServiceID(), client)
}

func (ch *ClientsImpl) Remove(id string) {
	ch.clients.Delete(id)
}

func (ch *ClientsImpl) Get(id string) (HopeIM.Client, bool) {
	if id == "" {
		logger.WithFields(logger.Fields{
			"module": "ClientsImpl",
		}).Error("client id is required")
	}

	val, ok := ch.clients.Load(id)
	if !ok {
		return nil, false
	}
	return val.(HopeIM.Client), true
}

// Services 返回服务列表, 可以传一对
func (ch *ClientsImpl) Services(kvs ...string) []HopeIM.Service {
	kvLen := len(kvs)
	if kvLen != 0 && kvLen != 2 {
		return nil
	}
	arr := make([]HopeIM.Service, 0)
	ch.clients.Range(func(key, val any) bool {
		ser := val.(HopeIM.Service)
		if kvLen > 0 && ser.GetMeta()[kvs[0]] != kvs[1] {
			return true
		}
		arr = append(arr, ser)
		return true

	})
	return arr
}
