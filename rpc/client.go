package rpc

import (
	"context"
	"encoding/json"
	"errors"
	"geek_micro/rpc/message"
	"net"
	"reflect"
	"time"

	"github.com/silenceper/pool"
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

				req := &message.Request{
					ServiceName: service.Name(),
					MethodName:  fieldTyp.Name,
					Data:        reqData,
				}

				req.SetHeadLength()
				req.SetBodyLength()

				// resp => eg: Response { data : []byte("{"Msg": "Hello, world"}") }
				resp, err := p.Invoke(ctx, req)

				if err != nil {
					// 这里可能是网络异常
					return []reflect.Value{retVal, reflect.ValueOf(err)}
				}

				var retErr error
				if len(resp.Error) > 0 {
					// 远端执行返回的错误
					retErr = errors.New(string(resp.Error))
				}

				if len(resp.Data) > 0 {
					// 返回值序列化
					err = json.Unmarshal(resp.Data, retVal.Interface())
					if err != nil {
						// 序列化出错
						return []reflect.Value{retVal, reflect.ValueOf(err)}
					}
				}

				var retErrVal reflect.Value
				if retErr == nil {
					retErrVal = reflect.Zero(reflect.TypeOf(new(error)).Elem())
				} else {
					retErrVal = reflect.ValueOf(retErr)
				}

				return []reflect.Value{retVal, retErrVal}
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

func (c *Client) Invoke(ctx context.Context, req *message.Request) (*message.Response, error) {
	// rpc通信中 传输需要进行
	data := message.EncodeReq(req)
	result, err := c.Send(data)
	if err != nil {
		return nil, err
	}
	return message.DecodeResp(result), nil
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

func (c *Client) Send(req []byte) ([]byte, error) {
	val, err := c.pool.Get()
	if err != nil {
		return nil, err
	}
	conn := val.(net.Conn)
	defer func() {
		_ = c.pool.Put(val)
	}()

	_, err = conn.Write(req)
	if err != nil {
		return nil, err
	}

	respBs, err := ReadMsg(conn)

	return respBs, err
}
