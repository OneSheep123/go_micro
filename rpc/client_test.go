package rpc

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/golang/mock/gomock"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSetStructFunc(t *testing.T) {
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
				data, _ := json.Marshal(&GetByIdReq{Id: 1})
				proxy.EXPECT().Invoke(gomock.Any(), &Request{
					ServiceName: "user-service",
					MethodName:  "GetById",
					Arg:         data,
				}).Return(&Response{
					data: []byte(`{"Msg":"hello, world"}`),
				}, nil)
				return proxy
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			err := setStructFunc(tt.service, tt.mock(ctrl))
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
