package container

import (
	"github.com/sjmshsh/HopeIM"
	"github.com/sjmshsh/HopeIM/wire/pkt"
)

type HashSelector struct{}

// Lookup a server
func (s *HashSelector) Lookup(header *pkt.Header, srvs []HopeIM.Service) string {
	ll := len(srvs)
	code := HashCode(header.ChannelId)
	return srvs[code%ll].ServiceID()
}
