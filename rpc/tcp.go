package rpc

import (
	"encoding/binary"
	"net"
)

const numOfLengthBytes = 8

func ReadMsg(conn net.Conn) ([]byte, error) {
	lengthByte := make([]byte, numOfLengthBytes)
	_, err := conn.Read(lengthByte)
	if err != nil {
		return nil, err
	}
	headerLength := binary.BigEndian.Uint32(lengthByte[:4])
	bodyLength := binary.BigEndian.Uint32(lengthByte[4:8])
	length := headerLength + bodyLength

	data := make([]byte, length)
	_, err = conn.Read(data[8:])
	if err != nil {
		return nil, err
	}
	copy(data[:8], lengthByte)
	return data, err
}
