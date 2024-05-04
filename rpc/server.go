package rpc

import (
	"context"
	"errors"
	"geek_micro/rpc/message"
	"geek_micro/rpc/serialize"
	"geek_micro/rpc/serialize/json"
	"net"
	"reflect"
)

type Serve struct {
	services map[string]reflectionStub
	// 服务端得支持多种序列化协议
	serializes map[uint8]serialize.Serialize
}

func NewServer() *Serve {
	res := &Serve{
		services:   make(map[string]reflectionStub, 16),
		serializes: make(map[uint8]serialize.Serialize, 4),
	}
	// 设置默认序列化协议
	s := &json.Serializer{}
	res.serializes[s.Code()] = s
	return res
}

func (s *Serve) RegisterSerialize(sl serialize.Serialize) {
	s.serializes[sl.Code()] = sl
}

func (s *Serve) RegisterService(service Service) {
	s.services[service.Name()] = reflectionStub{
		s:          service,
		value:      reflect.ValueOf(service),
		serializes: s.serializes,
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
		req := message.DecodeReq(data)

		resp, err := s.Invoke(context.Background(), req)
		// 这个你的业务 error
		if err != nil {
			// 所有的错误都在这里进行捕获塞入
			resp.Error = []byte(err.Error())
		}

		resp.SetHeadLength()
		resp.SetBodyLength()

		_, err = conn.Write(message.EncodeResp(resp))
		if err != nil {
			return err
		}
	}
}

func (s *Serve) Invoke(ctx context.Context, req *message.Request) (*message.Response, error) {
	resp := &message.Response{
		MessageId:  req.MessageId,
		Version:    req.Version,
		Compresser: req.Compresser,
		Serializer: req.Serializer,
	}
	service, ok := s.services[req.ServiceName]
	if !ok {
		return resp, errors.New("你要调用的服务不存在")
	}

	respData, err := service.invoke(ctx, req)
	resp.Data = respData

	return resp, err
}

type reflectionStub struct {
	s          Service
	value      reflect.Value
	serializes map[uint8]serialize.Serialize
}

func (s *reflectionStub) invoke(ctx context.Context, req *message.Request) ([]byte, error) {

	method := s.value.MethodByName(req.MethodName)
	in := make([]reflect.Value, 2)

	// in[0]：需要传入context
	in[0] = reflect.ValueOf(context.Background())

	// in[1]: GetByIdReq数据
	inReq := reflect.New(method.Type().In(1).Elem())
	serializer, ok := s.serializes[req.Serializer]
	if !ok {
		return nil, errors.New("micro: 不支持的序列化协议")
	}
	err := serializer.Decode(req.Data, inReq.Interface())

	if err != nil {
		return nil, err
	}

	in[1] = inReq
	result := method.Call(in)

	if result[1].Interface() != nil {
		// 执行返回的错误
		err = result[1].Interface().(error)
	}

	var res []byte
	if result[0].IsNil() {
		return nil, err
	} else {
		var er error
		res, er = serializer.Encode(result[0].Interface())
		if er != nil {
			return nil, er
		}
	}
	return res, err
}
