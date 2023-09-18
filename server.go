package HopeIM

import (
	"context"
	"net"
	"time"
)

const (
	DefaultReadWait  = time.Minute * 3
	DefaultWriteWait = time.Second * 10
	DefaultLoginWait = time.Second * 10
	DefaultHeartbeat = time.Second * 55
)

type Server interface {
	SetAcceptor(Acceptor)
	// SetMessageListener 设置上行消息监听器
	SetMessageListener(MessageListener)
	// SetStateListener 设置连接状态监听服务
	SetStateListener(StateListener)
	// SetReadWait 设置读超时
	SetReadWait(duration time.Duration)
	// SetChannelMap 设置Channel管理服务
	SetChannelMap(ChannelMap)
	// Start 用于在内部实现网络端口的监听和接收连接
	Start() error
	// Push 消息到指定的Channel当中
	// string channelID
	// []byte 序列化之后的消息数据
	Push(string, []byte) error
	// Shutdown 服务下线，关闭连接
	Shutdown(context.Context) error
}

type Acceptor interface {
	// Accept 返回一个握手完成的Channel对象或者一个error。
	// 业务层需要处理不同协议和网络环境的下连接握手协议
	Accept(conn Conn, duration time.Duration) (string, error)
}

// StateListener 设置一个状态监听器
type StateListener interface {
	Disconnect(string) error
}

type Channel interface {
	Conn
	Agent
	Close() error
	Readloop(lst MessageListener) error
	SetWriteWait(duration time.Duration)
	SetReadWait(duration time.Duration)
}

type MessageListener interface {
	Receive(Agent, []byte)
}

// Agent 表示发送方
type Agent interface {
	ID() string
	Push([]byte) error
}

type Conn interface {
	net.Conn
	ReadFrame() (Frame, error)
	WriteFrame(OpCode, []byte) error
	Flush() error
}

type Client interface {
	ID() string
	Name() string
	// Connect 主动向服务器地址发起连接
	Connect(string) error
	// SetDialer 设置一个拨号器，这个方法会在Connect当中被调用，完成连接的建立和握手
	SetDialer(Dialer)
	Send([]byte) error
	Read() (Frame, error)
	Close()
}

type Dialer interface {
	DialAndHandshake(DialerContext) (net.Conn, error)
}

type DialerContext struct {
	Id      string
	Name    string
	Address string
	Timeout time.Duration
}

type OpCode byte

const (
	OpBinary OpCode = 0x2
	OpClose  OpCode = 0x8
	OpPing   OpCode = 0x9
	OpPong   OpCode = 0xa
)

type Frame interface {
	SetOpCode(OpCode)
	GetOpCode() OpCode
	SetPayload([]byte)
	GetPayload() []byte
}
