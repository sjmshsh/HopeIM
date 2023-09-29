package HopeIM

import "github.com/sjmshsh/HopeIM/wire/pkt"

// Dispather defined a component how a message be dispatched to gateway
type Dispather interface {
	Push(gateway string, channels []string, p *pkt.LogicPkt) error
}
