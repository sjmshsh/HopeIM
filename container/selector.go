package container

import (
	"github.com/sjmshsh/HopeIM"
	"github.com/sjmshsh/HopeIM/wire/pkt"
	"hash/crc32"
)

func HashCode(key string) int {
	hash32 := crc32.NewIEEE()
	hash32.Write([]byte(key))
	return int(hash32.Sum32())
}

type Selector interface {
	Lookup(*pkt.Header, []HopeIM.Service) string
}
