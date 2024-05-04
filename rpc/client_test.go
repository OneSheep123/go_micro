package rpc

import (
	"context"
	"errors"
	"fmt"
	"geek_micro/rpc/message"
	"geek_micro/rpc/serialize/json"
	"github.com/golang/mock/gomock"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSetStructFunc(t *testing.T) {
	s := &json.Serializer{}
	tests := []struct {
		name    string
		service Service
		mock    func(ctrl *gomock.Controller) Proxy
		wantErr error
	}{
		{
			name:    "nil",
			service: nil,
			mock: func(ctrl *gomock.Controller) Proxy {
				return NewMockProxy(ctrl)
			},
			wantErr: errors.New("rpc: 不支持 nil"),
		},
		{
			name:    "*int",
			service: UserService{},
			mock: func(ctrl *gomock.Controller) Proxy {
				return NewMockProxy(ctrl)
			},
			wantErr: errors.New("rpc: 只支持指向结构体的一级指针"),
		},
		{
			name:    "ok",
			service: &UserService{},
			mock: func(ctrl *gomock.Controller) Proxy {
				proxy := NewMockProxy(ctrl)
				data, _ := s.Encode(&GetByIdReq{Id: 1})
				proxy.EXPECT().Invoke(gomock.Any(), &message.Request{
					ServiceName: "user-service",
					MethodName:  "GetById",
					Data:        data,
				}).Return(&message.Response{
					Data: []byte(`{"Msg":"hello, world"}`),
				}, nil)
				return proxy
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			err := setStructFunc(tt.service, tt.mock(ctrl), s)
			if err != nil {
				assert.Equal(t, tt.wantErr, err)
				return
			}
			resp, err := tt.service.(*UserService).GetById(context.Background(), &GetByIdReq{Id: 1})
			if err != nil {
				assert.Equal(t, tt.wantErr, err)
				return
			}
			fmt.Println(resp)
		})
	}
}
