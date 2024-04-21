package rpc

import (
	"context"
	"encoding/binary"
	"encoding/json"
	"errors"
	"net"
	"reflect"
	"time"
)

const numOfLengthBytes = 8

// InitClientProxy 要为 GetById 之类的函数类型的字段赋值
func InitClientProxy(address string, service Service) error {
	return setStructFunc(service, NewClient(address))
}

func setStructFunc(service Service, p Proxy) error {
	if service == nil {
		return errors.New("rpc: 不支持 nil")
	}
	vOf := reflect.ValueOf(service)
	tOf := vOf.Type()
	if tOf.Kind() != reflect.Pointer || tOf.Elem().Kind() != reflect.Struct {
		return errors.New("rpc: 只支持指向结构体的一级指针")
	}
	vOf = vOf.Elem()
	tOf = tOf.Elem()
	numField := vOf.NumField()
	for i := 0; i < numField; i++ {
		fieldVal := vOf.Field(i)
		fieldTyp := tOf.Field(i)

		if fieldVal.CanSet() {
			fn := func(args []reflect.Value) (results []reflect.Value) {
				//args[0] 是 context.Context
				//args[1] 是 req（用户的请求数据）
				ctx := args[0].Interface().(context.Context)

				// Out 对那个Type为函数类型时，第i+1个返回值
				// eg: GetByIdResp
				retVal := reflect.New(fieldTyp.Type.Out(0).Elem())

				reqData, err := json.Marshal(args[1].Interface())
				if err != nil {
					return []reflect.Value{retVal, reflect.ValueOf(err)}
				}

				req := &Request{
					ServiceName: service.Name(),
					MethodName:  fieldTyp.Name,
					Arg:         reqData,
				}

				// result => eg: Response { data : []byte("{"Msg": "Hello, world"}") }
				result, err := p.Invoke(ctx, req)
				if err != nil {
					return []reflect.Value{retVal, reflect.ValueOf(err)}
				}

				// 返回值序列化
				err = json.Unmarshal(result.data, retVal.Interface())
				if err != nil {
					return []reflect.Value{retVal, reflect.ValueOf(err)}
				}

				return []reflect.Value{retVal, reflect.Zero(reflect.TypeOf(new(error)).Elem())}
			}
			fnVal := reflect.MakeFunc(fieldTyp.Type, fn)
			fieldVal.Set(fnVal)
		}
	}
	return nil
}

type Client struct {
	Addr string
}

func (c *Client) Invoke(ctx context.Context, req *Request) (*Response, error) {
	// rpc通信中 传输需要进行序列化
	data, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}
	result, err := c.Send(data)
	if err != nil {
		return nil, err
	}
	return &Response{
		data: result,
	}, nil
}

func NewClient(addr string) *Client {
	return &Client{Addr: addr}
}

func (c *Client) Send(data []byte) ([]byte, error) {
	conn, err := net.DialTimeout("tcp", c.Addr, time.Second*3)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = conn.Close()
	}()

	reqLen := len(data)

	// 我要在这，构建请求数据
	// data = reqLen 的 64 位表示 + respData
	req := make([]byte, reqLen+numOfLengthBytes)
	// 第一步：
	// 把长度写进去前八个字节
	binary.BigEndian.PutUint64(req[:numOfLengthBytes], uint64(reqLen))
	// 第二步：
	// 写入数据
	copy(req[numOfLengthBytes:], data)

	_, err = conn.Write(req)
	if err != nil {
		return nil, err
	}

	lenBs := make([]byte, numOfLengthBytes)
	_, err = conn.Read(lenBs)
	if err != nil {
		return nil, err
	}

	// 我响应有多长？
	length := binary.BigEndian.Uint64(lenBs)

	respBs := make([]byte, length)
	_, err = conn.Read(respBs)
	if err != nil {
		return nil, err
	}

	return respBs, nil
}
