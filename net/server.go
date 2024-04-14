package net

import (
	"encoding/binary"
	"net"
)

const numOfLengthBytes = 8

type Serve struct {
}

func (s *Serve) Start(network, address string) error {
	listener, err := net.Listen(network, address)
	if err != nil {
		return err
	}
	for {
		conn, err := listener.Accept()
		if err != nil {
			return err
		}
		go func() {
			if err := s.handleConn(conn); err != nil {
				_ = conn.Close()
			}
		}()
	}
}

func (s *Serve) handleConn(conn net.Conn) error {
	for {
		lengthByte := make([]byte, numOfLengthBytes)
		_, err := conn.Read(lengthByte)
		if err != nil {
			return err
		}
		length := binary.BigEndian.Uint64(lengthByte)
		data := make([]byte, length)
		_, err = conn.Read(data)
		if err != nil {
			return err
		}
		resMsg := handleMsg(data)

		lenResMsg := len(resMsg)
		res := make([]byte, lenResMsg+numOfLengthBytes)
		binary.BigEndian.PutUint64(res[:numOfLengthBytes], uint64(lenResMsg))
		copy(res[numOfLengthBytes:], resMsg)

		_, err = conn.Write(res)
		if err != nil {
			return err
		}
	}
}

func handleMsg(req []byte) []byte {
	res := make([]byte, 2*len(req))
	copy(res[:len(req)], req)
	copy(res[len(req):], req)
	return res
}
