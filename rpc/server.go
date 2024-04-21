package rpc

import (
	"context"
	"encoding/json"
	"errors"
	"net"
	"reflect"
)

type Serve struct {
	services map[string]reflectionStub
}

func NewServer() *Serve {
	return &Serve{
		services: make(map[string]reflectionStub, 16),
	}
}

func (s *Serve) RegisterService(service Service) {
	s.services[service.Name()] = reflectionStub{
		s:     service,
		value: reflect.ValueOf(service),
	}
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
		data, err := ReadMsg(conn)
		if err != nil {
			return err
		}

		// 还原调用信息
		req := &Request{}
		err = json.Unmarshal(data, req)
		if err != nil {
			return err
		}

		resMsg, err := s.Invoke(context.Background(), req)
		// 这个可能你的业务 error
		// 暂时不知道怎么回传 error，所以我们简单记录一下
		if err != nil {
			return err
		}

		// 获取响应信息
		res := EncodeMsg(resMsg.data)

		_, err = conn.Write(res)
		if err != nil {
			return err
		}
	}
}

func (s *Serve) Invoke(ctx context.Context, req *Request) (*Response, error) {
	service, ok := s.services[req.ServiceName]
	if !ok {
		return nil, errors.New("你要调用的服务不存在")
	}
	resp, err := service.invoke(ctx, req.MethodName, req.Arg)
	if err != nil {
		return nil, err
	}
	return &Response{
		data: resp,
	}, nil
}

type reflectionStub struct {
	s     Service
	value reflect.Value
}

func (s *reflectionStub) invoke(ctx context.Context, methodName string, data []byte) ([]byte, error) {

	method := s.value.MethodByName(methodName)
	in := make([]reflect.Value, 2)

	// in[0]：需要传入context
	in[0] = reflect.ValueOf(context.Background())

	// in[1]: GetByIdReq数据
	inReq := reflect.New(method.Type().In(1).Elem())
	err := json.Unmarshal(data, inReq.Interface())

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
	return json.Marshal(result[0].Interface())
}
