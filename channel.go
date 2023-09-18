package HopeIM

import (
	"errors"
	"fmt"
	"github.com/sjmshsh/HopeIM/logger"
	"sync"
	"time"
)

type ChannelImpl struct {
	sync.Mutex
	id string
	Conn
	writeChan chan []byte
	once      sync.Once
	writeWait time.Duration
	readWait  time.Duration
	closed    *Event
}

func NewChannel(id string, conn Conn) Channel {
	log := logger.WithFields(logger.Fields{
		"module": "channel",
		"id":     id,
	})

	ch := &ChannelImpl{
		id:        id,
		Conn:      conn,
		writeChan: make(chan []byte, 5),
		writeWait: DefaultWriteWait,
		readWait:  DefaultReadWait,
		closed:    NewEvent(),
	}

	go func() {
		err := ch.writeloop()
		if err != nil {
			log.Info(err)
		}
	}()

	return ch
}

func (ch *ChannelImpl) writeloop() error {
	for {
		select {
		case payload := <-ch.writeChan:
			err := ch.WriteFrame(OpBinary, payload)
			if err != nil {
				return err
			}
			chanLen := len(ch.writeChan)
			for i := 0; i < chanLen; i++ {
				payload = <-ch.writeChan
				err := ch.WriteFrame(OpBinary, payload)
				if err != nil {
					return err
				}
			}
			err = ch.Conn.Flush()
			if err != nil {
				return err
			}
		case <-ch.closed.Done():
			return nil
		}
	}
}

func (ch *ChannelImpl) ID() string { return ch.id }

// Push 异步写数据
func (ch *ChannelImpl) Push(payload []byte) error {
	if ch.closed.HasFired() {
		return fmt.Errorf("channel %s has closed", ch.id)
	}
	// 异步写
	ch.writeChan <- payload
	return nil
}

func (ch *ChannelImpl) WriteFrame(code OpCode, payload []byte) error {
	_ = ch.Conn.SetWriteDeadline(time.Now().Add(ch.writeWait))
	return ch.Conn.WriteFrame(code, payload)
}

func (ch *ChannelImpl) Close() error {
	ch.once.Do(func() {
		close(ch.writeChan)
		ch.closed.Fire()
	})
	return nil
}

// SetWriteWait 设置写超时
func (ch *ChannelImpl) SetWriteWait(writeWait time.Duration) {
	if writeWait == 0 {
		return
	}
	ch.writeWait = writeWait
}

func (ch *ChannelImpl) SetReadWait(readwait time.Duration) {
	if readwait == 0 {
		return
	}
	ch.writeWait = readwait
}

func (ch *ChannelImpl) Readloop(lst MessageListener) error {
	ch.Lock()
	defer ch.Unlock()
	log := logger.WithFields(logger.Fields{
		"struct": "ChannelImpl",
		"func":   "Readloop",
		"id":     ch.id,
	})
	for {
		_ = ch.SetReadDeadline(time.Now().Add(ch.readWait))

		frame, err := ch.ReadFrame()
		if err != nil {
			return err
		}
		if frame.GetOpCode() == OpClose {
			return errors.New("remote side close the channel")
		}
		if frame.GetOpCode() == OpPing {
			log.Trace("recv a ping; resp with a pong")
			_ = ch.WriteFrame(OpPong, nil)
			continue
		}
		payload := frame.GetPayload()
		if len(payload) == 0 {
			continue
		}
		go lst.Receive(ch, payload)
	}
}
