package rpc

import (
	"context"
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
			wantErr: errors.New("rpc: 不支持 nil"),
		},
		{
			name:    "*int",
			service: UserService{},
			wantErr: errors.New("rpc: 只支持指向结构体的一级指针"),
		},
		{
			name:    "ok",
			service: &UserService{},
			mock: func(ctrl *gomock.Controller) Proxy {
				proxy := NewMockProxy(ctrl)
				proxy.EXPECT().Invoke(gomock.Any(), &Request{
					ServiceName: "user-service",
					MethodName:  "GetById",
					Arg:         &GetByIdReq{Id: 1},
				}).Return(&Response{}, nil)
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
			}
			resp, err := tt.service.(*UserService).GetById(context.Background(), &GetByIdReq{Id: 1})
			if err != nil {
				assert.Equal(t, tt.wantErr, err)
			}
			fmt.Println(resp)
		})
	}
}
