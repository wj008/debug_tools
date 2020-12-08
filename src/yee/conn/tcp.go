package conn

import (
	"encoding/binary"
	"errors"
	"fmt"
	"log"
	"net"
	"time"
)

type TcpConn struct {
	net.Conn
	Id  string
	typ string
}

type TcpListener struct {
	net.Addr
	Clients chan *TcpConn
}

func TcpListen(addr, typ string) (lst *TcpListener, err error) {
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return
	}
	lst = &TcpListener{
		Addr:    listener.Addr(),
		Clients: make(chan *TcpConn),
	}
	go func() {
		for {
			rawConn, err := listener.Accept()
			if err != nil {
				//log.Printf("Failed to accept new UDP connection of type %s: %v", typ, err)
				continue
			}
			rmAddr := rawConn.RemoteAddr()
			c := wrapConn(rawConn, rmAddr.String(), typ)
			//log.Printf("New connection from %v", rmAddr)
			lst.Clients <- c
		}
	}()
	return
}

func TcpDial(addr string) (conn *TcpConn, err error) {
	rawConn, err := net.Dial("tcp", addr)
	if err != nil {
		return
	}
	rmAddr := rawConn.RemoteAddr()
	conn = wrapConn(rawConn, rmAddr.String(), "cli")
	return
}

/**
关闭
*/
func (c *TcpConn) Close() (err error) {
	if err := c.Conn.Close(); err == nil {
		log.Println("Closing")
	}
	return
}

/**
读取消息
*/
func (c *TcpConn) ReadMsg() (buffer []byte, typ int32, err error) {
	err = binary.Read(c, binary.LittleEndian, &typ)
	if err != nil {
		return
	}
	var sz int32
	err = binary.Read(c, binary.LittleEndian, &sz)
	if err != nil {
		return
	}
	iz := int(sz)
	if iz == 0 {
		buffer = make([]byte, 0)
		return
	}
	buffer = make([]byte, iz)
	temp := buffer[0:iz]
	retry := 0
	nLen := 0
	for {
		retry++
		if retry > 1000 {
			err = errors.New(fmt.Sprintf("Expected to read %d bytes, but only read %d", sz, nLen))
			return
		}
		n, err1 := c.Read(temp)
		if err1 != nil {
			err = err1
			return
		}
		nLen += n
		if n < iz {
			temp = buffer[n:iz]
			iz = iz - n
			continue
		} else {
			break
		}
	}
	return
}

/**
写入消息
*/
func (c *TcpConn) WriteMsg(buffer []byte, typ int32) (err error) {
	err = binary.Write(c, binary.LittleEndian, typ)
	if err != nil {
		return
	}
	l := 0
	if buffer != nil {
		l = len(buffer)
	}
	err = binary.Write(c, binary.LittleEndian, int32(l))
	if err != nil {
		return
	}
	if l == 0 {
		return
	}
	c.SetWriteDeadline(time.Time{})
	if _, err = c.Write(buffer); err != nil {
		return
	}
	return nil
}

/**
包装链接
*/
func wrapConn(conn net.Conn, id string, typ string) *TcpConn {
	switch c := conn.(type) {
	case *TcpConn:
		return c
	case *net.TCPConn:
		return &TcpConn{conn, id, typ}
	}
	return nil
}
