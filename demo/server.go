package main

import (
	"errors"
	"fmt"
	"github.com/gobwas/ws"
	"github.com/sirupsen/logrus"
	"net"
	"net/http"
	"sync"
)

// Server is a websocket Server
type Server struct {
	once    sync.Once
	id      string
	address string
	sync.Mutex
	// 会话列表
	users map[string]net.Conn
}

// NewServer NewServer
func NewServer(id, address string) *Server {
	return newServer(id, address)
}

func newServer(id, address string) *Server {
	return &Server{
		id:      id,
		address: address,
		users:   make(map[string]net.Conn, 100),
	}
}

func (s *Server) Start() error {
	mux := http.NewServeMux()
	log := logrus.WithFields(logrus.Fields{
		"module": "Server",
		"listen": s.address,
		"id":     s.id,
	})

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// step1. 升级
		conn, _, _, err := ws.UpgradeHTTP(r, w)
		if err != nil {
			conn.Close()
			return
		}
		// step2. 读取userId
		user := r.URL.Query().Get("user")
		if user == "" {
			conn.Close()
			return
		}
		// step3. 添加到会话管理中
		old, ok := s.addUser(user, conn)
		if ok {
			// 断开旧的连接
			old.Close()
		}
		log.Infof("user %s in", user)

		go func(user string, conn net.Conn) {
			// step4. 读取消息
			err := s.readloop(user, conn)
			if err != nil {
				log.Error(err)
			}
			conn.Close()
			// step5. 连接断开，删除用户
			s.delUser(user)

			log.Infof("connection of %s closed", user)
		}(user, conn)
	})
	log.Infoln("started")
	return http.ListenAndServe(s.address, mux)
}

func (s *Server) addUser(user string, conn net.Conn) (net.Conn, bool) {
	s.Lock()
	defer s.Unlock()
	// 返回旧的连接
	old, ok := s.users[user]
	s.users[user] = conn
	return old, ok
}

func (s *Server) delUser(user string) {
	s.Lock()
	defer s.Unlock()
	delete(s.users, user)
}

func (s *Server) Shutdown() {
	s.once.Do(func() {
		s.Lock()
		defer s.Unlock()
		for _, conn := range s.users {
			conn.Close()
		}
	})
}

func (s *Server) readloop(user string, conn net.Conn) error {
	for {
		// 从TCP缓冲中读取一帧数据
		frame, err := ws.ReadFrame(conn)
		if err != nil {
			return err
		}
		if frame.Header.OpCode == ws.OpClose {
			return errors.New("remote side close the conn")
		}

		if frame.Header.Masked {
			// 使用Mask解码数据包，由于websocket协议规定客户端发送数据时，必须使用一个随机的mask值对消息体做一次编码
			// 因此在服务端旧需要解码，否则内容是乱的
			ws.Cipher(frame.Payload, frame.Header.Mask, 0)
		}
		// 接收文本帧的内容
		if frame.Header.OpCode == ws.OpText {
			go s.handle(user, string(frame.Payload))
		}
	}
}

// 广播消息
func (s *Server) handle(user string, message string) {
	logrus.Infof("recv message %s from %s", message, user)
	s.Lock()
	defer s.Unlock()
	broadcast := fmt.Sprintf("%s -- FROM %s", message, user)
	for u, conn := range s.users {
		// 不发给自己
		if u == user {
			continue
		}
		logrus.Infof("send to %s: %s", u, broadcast)
		err := s.writeText(conn, broadcast)
		if err != nil {
			logrus.Errorf("write to %s failed, error: %v", user, err)
		}
	}
}

func (s *Server) writeText(conn net.Conn, message string) error {
	// 创建文本帧数据
	f := ws.NewTextFrame([]byte(message))
	return ws.WriteFrame(conn, f)
}
