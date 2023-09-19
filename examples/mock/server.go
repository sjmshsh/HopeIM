package mock

import (
	"github.com/sjmshsh/HopeIM"
	"github.com/sjmshsh/HopeIM/logger"
	"github.com/sjmshsh/HopeIM/naming"
	"github.com/sjmshsh/HopeIM/tcp"
	"github.com/sjmshsh/HopeIM/websocket"
	"time"
)

type ServerDemo struct{}

func (s *ServerDemo) Start(id, protocol, addr string) {
	var srv HopeIM.Server
	service := &naming.DefaultService{
		Id:       id,
		Protocol: protocol,
	}

	if protocol == "ws" {
		srv = websocket.NewServer(addr, service)
	} else if protocol == "tcp" {
		srv = tcp.NewServer(addr, service)
	}

	handler := &ServerHandler{}
	srv.SetReadWait(time.Minute)
	srv.SetAcceptor(handler)
	srv.SetMessageListener(handler)
	srv.SetStateListener(handler)

	err := srv.Start()
	if err != nil {
		panic(err)
	}
}

type ServerHandler struct{}

func (h *ServerHandler) Accept(conn HopeIM.Conn, timeout time.Duration) (string, error) {
	// 1. 读取: 客户端发送的鉴权数据包
	frame, err := conn.ReadFrame()
	if err != nil {
		return "", err
	}
	// 2. 解析：数据包内容就是userId
	userID := string(frame.GetPayload())
	// 3. 鉴权：这里只是为了实例做一个fake验证，非空
	if userID == "" {
		return "", err
	}
	return userID, nil
}

func (h *ServerHandler) Receive(ag HopeIM.Agent, payload []byte) {
	ack := string(payload) + " from server"
	_ = ag.Push([]byte(ack))
}

func (h *ServerHandler) Disconnect(id string) error {
	logger.Warnf("disconnect %s", id)
	return nil
}
