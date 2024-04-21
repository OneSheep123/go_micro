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
	length := binary.BigEndian.Uint64(lengthByte)
	data := make([]byte, length)
	_, err = conn.Read(data)
	if err != nil {
		return nil, err
	}
	return data, err
}

func EncodeMsg(resMsg []byte) []byte {
	lenResMsg := len(resMsg)
	res := make([]byte, lenResMsg+numOfLengthBytes)
	binary.BigEndian.PutUint64(res[:numOfLengthBytes], uint64(lenResMsg))
	copy(res[numOfLengthBytes:], resMsg)
	return res
}
