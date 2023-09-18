package tcp

import (
	"bufio"
	"github.com/sjmshsh/HopeIM"
	"github.com/sjmshsh/HopeIM/wire/endian"
	"io"
	"net"
)

type Frame struct {
	OpCode  HopeIM.OpCode
	Payload []byte
}

// SetOpCode SetOpCode
func (f *Frame) SetOpCode(code HopeIM.OpCode) {
	f.OpCode = code
}

// GetOpCode GetOpCode
func (f *Frame) GetOpCode() HopeIM.OpCode {
	return f.OpCode
}

// SetPayload SetPayload
func (f *Frame) SetPayload(payload []byte) {
	f.Payload = payload
}

// GetPayload GetPayload
func (f *Frame) GetPayload() []byte {
	return f.Payload
}

type TcpConn struct {
	net.Conn
	rd *bufio.Reader
	wr *bufio.Writer
}

func NewConn(conn net.Conn) HopeIM.Conn {
	return &TcpConn{
		Conn: conn,
		rd:   bufio.NewReaderSize(conn, 4096),
		wr:   bufio.NewWriterSize(conn, 1024),
	}
}

func NewConnWithRW(conn net.Conn, rd *bufio.Reader, wr *bufio.Writer) *TcpConn {
	return &TcpConn{
		Conn: conn,
		rd:   rd,
		wr:   wr,
	}
}

func (c *TcpConn) Flush() error {
	return c.wr.Flush()
}

func WriteFrame(w io.Writer, code HopeIM.OpCode, payload []byte) error {
	if err := endian.WriteUint8(w, uint8(code)); err != nil {
		return err
	}
	if err := endian.WriteBytes(w, payload); err != nil {
		return err
	}
	return nil
}

func (c *TcpConn) WriteFrame(code HopeIM.OpCode, payload []byte) error {
	return WriteFrame(c.wr, code, payload)
}

func (c *TcpConn) ReadFrame() (HopeIM.Frame, error) {
	opcode, err := endian.ReadUint8(c.rd)
	if err != nil {
		return nil, err
	}
	payload, err := endian.ReadBytes(c.rd)
	if err != nil {
		return nil, err
	}
	return &Frame{
		OpCode:  HopeIM.OpCode(opcode),
		Payload: payload,
	}, nil
}
