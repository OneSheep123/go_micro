package rpc

import (
	"context"
	"errors"
	"fmt"
	"reflect"
)

// InitClientProxy 要为 GetById 之类的函数类型的字段赋值
func InitClientProxy(service Service) error {
	return nil
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
				req := &Request{
					ServiceName: service.Name(),
					MethodName:  fieldTyp.Name,
					Arg:         args[1].Interface(),
				}

				// Out 对那个Type为函数类型时，第i+1个返回值
				retVal := reflect.New(fieldTyp.Type.Out(0)).Elem()

				res, err := p.Invoke(ctx, req)
				if err != nil {
					return []reflect.Value{retVal, reflect.ValueOf(err)}
				}

				fmt.Println(res)
				// todo: 调整为res
				return []reflect.Value{retVal, reflect.Zero(reflect.TypeOf(new(error)).Elem())}
			}
			fnVal := reflect.MakeFunc(fieldTyp.Type, fn)
			fieldVal.Set(fnVal)
		}
	}
	return nil
}
