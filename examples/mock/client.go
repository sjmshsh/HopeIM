package mock

import (
	"context"
	"github.com/gobwas/ws"
	"github.com/gobwas/ws/wsutil"
	"github.com/klintcheng/kim"
	"github.com/sjmshsh/HopeIM"
	"github.com/sjmshsh/HopeIM/logger"
	"github.com/sjmshsh/HopeIM/tcp"
	"github.com/sjmshsh/HopeIM/websocket"
	"net"
	"time"
)

type ClientDemo struct{}

func (c *ClientDemo) Start(userID, protocol, addr string) {
	var cli HopeIM.Client

	// step1: 初始化客户端
	if protocol == "ws" {
		cli := websocket.NewClient(userID, "client", websocket.ClientOptions{})
		// set dialer
		cli.SetDialer(&WebsocketDialer{})
	} else if protocol == "tcp" {
		cli = tcp.NewClient("test1", "client", tcp.ClientOptions{})
		cli.SetDialer(&TCPDialer{})
	}

	// step2: 建立连接
	err := cli.Connect(addr)
	if err != nil {
		logger.Error(err)
	}

	count := 5
	go func() {
		// step3: 发送消息然后退出
		for i := 0; i < count; i++ {
			err := cli.Send([]byte("hello"))
			if err != nil {
				logger.Error(err)
				return
			}
			time.Sleep(time.Millisecond * 10)
		}
	}()

	// step4: 接收消息
	recv := 0
	for {
		frame, err := cli.Read()
		if err != nil {
			logger.Infoln(err)
			break
		}
		if frame.GetOpCode() != HopeIM.OpBinary {
			continue
		}
		recv++
		logger.Warnf("%s receive message [%s]", cli.ID(), frame.GetPayload())
		if recv == count { // 接收完消息
			break
		}
	}
	cli.Close()
}

type ClientHandler struct {
}

// Receive default listener
func (h *ClientHandler) Receive(ag kim.Agent, payload []byte) {
	logger.Warnf("%s receive message [%s]", ag.ID(), string(payload))
}

// Disconnect default listener
func (h *ClientHandler) Disconnect(id string) error {
	logger.Warnf("disconnect %s", id)
	return nil
}

type WebsocketDialer struct{}

func (d *WebsocketDialer) DialAndHandshake(ctx HopeIM.DialerContext) (net.Conn, error) {
	// 1. 调用ws.Dial拨号
	conn, _, _, err := ws.Dial(context.TODO(), ctx.Address)
	if err != nil {
		return nil, err
	}
	// 2. 发送用户认证信息，实例就是userID
	err = wsutil.WriteClientBinary(conn, []byte(ctx.Id))
	if err != nil {
		return nil, err
	}
	return conn, nil
}

// TCPDialer TCPDialer
type TCPDialer struct {
}

// DialAndHandshake DialAndHandshake
func (d *TCPDialer) DialAndHandshake(ctx HopeIM.DialerContext) (net.Conn, error) {
	logger.Info("start dial: ", ctx.Address)
	// 1 调用net.Dial拨号
	conn, err := net.DialTimeout("tcp", ctx.Address, ctx.Timeout)
	if err != nil {
		return nil, err
	}
	// 2. 发送用户认证信息，示例就是userid
	err = tcp.WriteFrame(conn, HopeIM.OpBinary, []byte(ctx.Id))
	if err != nil {
		return nil, err
	}
	// 3. return conn
	return conn, nil
}
