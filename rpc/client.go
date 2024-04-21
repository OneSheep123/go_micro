package rpc

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/silenceper/pool"
	"net"
	"reflect"
	"time"
)

// InitClientProxy 要为 GetById 之类的函数类型的字段赋值
func InitClientProxy(address string, service Service) error {
	c, err := NewClient(address)
	if err != nil {
		return err
	}

	return setStructFunc(service, c)
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
	pool pool.Pool
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

func NewClient(addr string) (*Client, error) {
	p, err := pool.NewChannelPool(&pool.Config{
		InitialCap:  1,
		MaxCap:      30,
		MaxIdle:     10,
		IdleTimeout: time.Minute,
		Factory: func() (interface{}, error) {
			conn, err := net.DialTimeout("tcp", addr, time.Second*3)
			if err != nil {
				return nil, err
			}
			return conn, nil
		},
		Close: func(i interface{}) error {
			return i.(net.Conn).Close()
		},
	})
	if err != nil {
		return nil, err
	}
	return &Client{pool: p}, nil
}

func (c *Client) Send(data []byte) ([]byte, error) {
	val, err := c.pool.Get()
	if err != nil {
		return nil, err
	}
	conn := val.(net.Conn)
	defer func() {
		_ = c.pool.Put(val)
	}()

	res := EncodeMsg(data)
	_, err = conn.Write(res)
	if err != nil {
		return nil, err
	}

	respBs, err := ReadMsg(conn)

	return respBs, err
}
