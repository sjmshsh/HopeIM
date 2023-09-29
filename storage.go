package HopeIM

import (
	"errors"
	"github.com/sjmshsh/HopeIM/wire/pkt"
)

var ErrSessionNil = errors.New("err:session nil")

type SessionStorage interface {
	Add(session *pkt.Session) error
	Delete(account string, channelId string) error
	Get(channelId string) (*pkt.Session, error)
	GetLocations(account ...string) ([]*Location, error)
	GetLocation(account string, device string) (*Location, error)
}
