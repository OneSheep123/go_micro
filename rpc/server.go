package rpc

import (
	"context"
	"encoding/binary"
	"encoding/json"
	"errors"
	"net"
	"reflect"
)

type Serve struct {
	services map[string]Service
}

func NewServer() *Serve {
	return &Serve{
		services: make(map[string]Service, 16),
	}
}

func (s *Serve) RegisterService(service Service) {
	s.services[service.Name()] = service
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
		resMsg, err := s.handleMsg(data)
		// 这个可能你的业务 error
		// 暂时不知道怎么回传 error，所以我们简单记录一下
		if err != nil {
			return err
		}
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

func (s *Serve) handleMsg(reqData []byte) ([]byte, error) {

	req := &Request{}
	err := json.Unmarshal(reqData, req)
	if err != nil {
		return nil, err
	}
	service, ok := s.services[req.ServiceName]
	if !ok {
		return nil, errors.New("你要调用的服务不存在")
	}
	vOf := reflect.ValueOf(service)
	method := vOf.MethodByName(req.MethodName)
	in := make([]reflect.Value, 2)

	// in[0]：需要传入context
	in[0] = reflect.ValueOf(context.Background())

	// in[1]: GetByIdReq数据
	inReq := reflect.New(method.Type().In(1).Elem())
	err = json.Unmarshal(req.Arg, inReq.Interface())
	if err != nil {
		return nil, err
	}

	in[1] = inReq
	result := method.Call(in)

	// result[0]是返回值(eg: GetByIdResp)
	// result[1]是error
	if result[1].Interface() != nil {
		return nil, result[1].Interface().(error)
	}

	// 返回时候也需要进行序列化一下
	res, err := json.Marshal(result[0].Interface())
	if err != nil {
		return nil, err
	}
	return res, nil
}
