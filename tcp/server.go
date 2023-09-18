package tcp

import (
	"context"
	"errors"
	"fmt"
	"github.com/segmentio/ksuid"
	"github.com/sjmshsh/HopeIM"
	"github.com/sjmshsh/HopeIM/logger"
	"github.com/sjmshsh/HopeIM/naming"
	"net"
	"sync"
	"time"
)

type ServerOptions struct {
	loginwait time.Duration // 登录超时
	readwait  time.Duration // 读超时
	writewait time.Duration // 读超时
}

type Server struct {
	listen string
	naming.ServiceRegistration
	HopeIM.ChannelMap
	HopeIM.Acceptor
	HopeIM.MessageListener
	HopeIM.StateListener
	once    sync.Once
	options ServerOptions
	quit    *HopeIM.Event
}

func NewServer(listen string, service naming.ServiceRegistration) HopeIM.Server {
	return &Server{
		listen:              listen,
		ServiceRegistration: service,
		ChannelMap:          HopeIM.NewChannels(100),
		quit:                HopeIM.NewEvent(),
		options: ServerOptions{
			loginwait: HopeIM.DefaultLoginWait,
			readwait:  HopeIM.DefaultReadWait,
			writewait: time.Second * 10,
		},
	}
}

func (s *Server) Start() error {
	log := logger.WithFields(logger.Fields{
		"module": "tcp.server",
		"listen": s.listen,
		"id":     s.ServiceID(),
	})

	if s.StateListener == nil {
		return fmt.Errorf("StateListener is nil")
	}
	if s.Acceptor == nil {
		s.Acceptor = new(defaultAcceptor)
	}

	lst, err := net.Listen("tcp", s.listen)
	if err != nil {
		return err
	}
	log.Info("started")
	for {
		rawconn, err := lst.Accept()
		if err != nil {
			rawconn.Close()
			log.Warn(err)
			continue
		}
		go func(rawconn net.Conn) {
			conn := NewConn(rawconn)

			id, err := s.Accept(conn, s.options.loginwait)
			if err != nil {
				_ = conn.WriteFrame(HopeIM.OpClose, []byte(err.Error()))
				conn.Close()
				return
			}
			if _, ok := s.Get(id); ok {
				log.Warnf("channel %s existed", id)
				_ = conn.WriteFrame(HopeIM.OpClose, []byte("channelId is repeated"))
				conn.Close()
				return
			}

			channel := HopeIM.NewChannel(id, conn)
			channel.SetReadWait(s.options.readwait)
			channel.SetWriteWait(s.options.writewait)

			s.Add(channel)

			log.Info("accept ", channel)
			err = channel.Readloop(s.MessageListener)
			if err != nil {
				log.Info(err)
			}
			s.Remove(channel.ID())
			_ = s.Disconnect(channel.ID())
			channel.Close()
		}(rawconn)

		select {
		case <-s.quit.Done():
			return fmt.Errorf("listen exited")
		default:
		}
	}
}

func (s *Server) Shutdown(ctx context.Context) error {
	log := logger.WithFields(logger.Fields{
		"module": "tcp.server",
		"id":     s.ServiceID(),
	})
	s.once.Do(func() {
		defer func() {
			log.Infoln("shutdown")
		}()
		channels := s.ChannelMap.All()
		for _, ch := range channels {
			ch.Close()

			select {
			case <-ctx.Done():
				return
			default:
				continue
			}
		}
	})
	return nil
}

func (s *Server) Push(id string, data []byte) error {
	ch, ok := s.ChannelMap.Get(id)
	if !ok {
		return errors.New("channel no found")
	}
	return ch.Push(data)
}

func (s *Server) SetChannelMap(channels HopeIM.ChannelMap) {
	s.ChannelMap = channels
}

func (s *Server) SetReadWait(readwait time.Duration) {
	s.options.readwait = readwait
}

func (s *Server) SetStateListener(listener HopeIM.StateListener) {
	s.StateListener = listener
}

func (s *Server) SetMessageListener(listener HopeIM.MessageListener) {
	s.MessageListener = listener
}

func (s *Server) SetAcceptor(acceptor HopeIM.Acceptor) {
	s.Acceptor = acceptor
}

type defaultAcceptor struct{}

// Accept defaultAcceptor
func (a *defaultAcceptor) Accept(conn HopeIM.Conn, timeout time.Duration) (string, error) {
	return ksuid.New().String(), nil
}
