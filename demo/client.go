package main

import (
	"context"
	"errors"
	"github.com/gobwas/ws"
	"github.com/sirupsen/logrus"
	"net"
	"net/url"
)

type handler struct {
	conn  net.Conn
	close chan struct{}
	recv  chan []byte
}

func (h handler) readloop(conn net.Conn) error {
	logrus.Info("readloop started")
	for {
		frame, err := ws.ReadFrame(conn)
		if err != nil {
			return err
		}
		if frame.Header.OpCode == ws.OpClose {
			return errors.New("remote side close the channel")
		}
		if frame.Header.OpCode == ws.OpText {
			h.recv <- frame.Payload
		}
	}
}

func connect(addr string) (*handler, error) {
	_, err := url.Parse(addr)
	if err != nil {
		return nil, err
	}

	conn, _, _, err := ws.Dial(context.Background(), addr)
	if err != nil {
		return nil, err
	}

	h := handler{
		conn:  conn,
		close: make(chan struct{}, 1),
		recv:  make(chan []byte, 10),
	}

	go func() {
		err := h.readloop(conn)
		if err != nil {
			logrus.Warn(err)
		}
		// 通知上层
		h.close <- struct{}{}
	}()

	return &h, nil
}
