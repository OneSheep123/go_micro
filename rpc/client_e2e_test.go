package rpc

import (
	"context"
	"errors"
	"geek_micro/rpc/proto/gen"
	"geek_micro/rpc/serialize/proto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func TestInitClientProto(t *testing.T) {
	// 初始化服务端
	server := NewServer()
	service := &UserServiceServer{}
	// 服务端注册方法
	server.RegisterService(service)
	server.RegisterSerialize(&proto.Serializer{})
	go func() {
		err := server.Start("tcp", ":8081")
		t.Log(err)
	}()
	time.Sleep(time.Second * 3)

	// 初始化客户端
	us := &UserService{}
	client, err := NewClient("localhost:8081", ClientWithSerializer(&proto.Serializer{}))
	require.NoError(t, err)
	err = client.InitService(us)
	require.NoError(t, err)

	testCases := []struct {
		name string
		mock func()

		wantErr  error
		wantResp *GetByIdResp
	}{
		{
			name: "no error",
			mock: func() {
				service.Err = nil
				service.Msg = "hello, world"
			},
			wantResp: &GetByIdResp{
				Msg: "hello, world",
			},
		},
		{
			name: "error",
			mock: func() {
				service.Err = errors.New("error")
				service.Msg = ""
			},
			wantErr:  errors.New("error"),
			wantResp: &GetByIdResp{},
		},
		{
			name: "error and msg",
			mock: func() {
				service.Err = errors.New("error")
				service.Msg = "123"
			},
			wantErr: errors.New("error"),
			wantResp: &GetByIdResp{
				Msg: "123",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.mock()
			resp, er := us.GetByIdProto(context.Background(), &gen.GetByIdReq{Id: 123})
			assert.Equal(t, tc.wantErr, er)
			if resp != nil && resp.User != nil {
				assert.Equal(t, tc.wantResp.Msg, resp.User.Name)
			}
		})
	}

}

func TestInitClientJSON(t *testing.T) {
	// 初始化服务端
	server := NewServer()
	service := &UserServiceServer{}
	// 服务端注册方法
	server.RegisterService(service)
	go func() {
		err := server.Start("tcp", ":8081")
		t.Log(err)
	}()
	time.Sleep(time.Second * 3)

	// 初始化客户端
	us := &UserService{}
	client, err := NewClient("localhost:8081")
	require.NoError(t, err)
	err = client.InitService(us)
	require.NoError(t, err)

	testCases := []struct {
		name string
		mock func()

		wantErr  error
		wantResp *GetByIdResp
	}{
		{
			name: "no error",
			mock: func() {
				service.Err = nil
				service.Msg = "hello, world"
			},
			wantResp: &GetByIdResp{
				Msg: "hello, world",
			},
		},
		{
			name: "error",
			mock: func() {
				service.Err = errors.New("error")
				service.Msg = ""
			},
			wantErr:  errors.New("error"),
			wantResp: &GetByIdResp{},
		},
		{
			name: "error and msg",
			mock: func() {
				service.Err = errors.New("error")
				service.Msg = "123"
			},
			wantErr: errors.New("error"),
			wantResp: &GetByIdResp{
				Msg: "123",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.mock()
			resp, er := us.GetById(context.Background(), &GetByIdReq{Id: 123})
			assert.Equal(t, tc.wantErr, er)
			assert.Equal(t, tc.wantResp, resp)
		})
	}

}
