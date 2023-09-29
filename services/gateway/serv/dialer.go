package serv

import (
	"github.com/golang/protobuf/proto"
	"github.com/sjmshsh/HopeIM"
	"github.com/sjmshsh/HopeIM/logger"
	"github.com/sjmshsh/HopeIM/tcp"
	"github.com/sjmshsh/HopeIM/wire/pkt"
	"net"
)

type TcpDialer struct {
	ServiceId string
}

func NewDialer(serviceId string) HopeIM.Dialer {
	return &TcpDialer{
		ServiceId: serviceId,
	}
}

func (d *TcpDialer) DialAndHandshake(ctx HopeIM.DialerContext) (net.Conn, error) {
	// 1. 拨号建立连接
	conn, err := net.DialTimeout("tcp", ctx.Address, ctx.Timeout)
	if err != nil {
		return nil, err
	}
	req := &pkt.InnerHandshakeReq{
		ServiceId: d.ServiceId,
	}
	logger.Infof("send req %v", req)
	// 2. 把自己的ServiceId发送给对方
	bts, _ := proto.Marshal(req)
	err = tcp.WriteFrame(conn, HopeIM.OpBinary, bts)
	if err != nil {
		return nil, err
	}
	return conn, err
}
