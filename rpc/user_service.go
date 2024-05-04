package rpc

import (
	"context"
	"geek_micro/rpc/proto/gen"
	"log"
)

type UserService struct {
	// 用反射来赋值
	// 类型是函数的字段，它不是方法（它不是定义在 UserService 上的方法）
	// 本质上是一个字段
	GetById func(ctx context.Context, req *GetByIdReq) (*GetByIdResp, error)

	GetByIdProto func(ctx context.Context, req *gen.GetByIdReq) (*gen.GetByIdResp, error)
}

func (u UserService) Name() string {
	return "user-service"
}

type GetByIdReq struct {
	Id int
}

type GetByIdResp struct {
	Msg string
}

// UserServiceServer 业务实际上实现的方法
type UserServiceServer struct {
	Msg string
	Err error
}

func (u *UserServiceServer) GetById(ctx context.Context, req *GetByIdReq) (*GetByIdResp, error) {
	log.Printf("[UserServiceServer.GetById] 接收到远程的请求数据为%v\n", req)
	return &GetByIdResp{
		Msg: u.Msg,
	}, u.Err
}

func (u *UserServiceServer) GetByIdProto(ctx context.Context, req *gen.GetByIdReq) (*gen.GetByIdResp, error) {
	log.Printf("[UserServiceServer.GetById] 接收到远程的请求数据为%v\n", req)
	return &gen.GetByIdResp{
		User: &gen.User{
			Name: u.Msg,
		},
	}, u.Err
}

func (u *UserServiceServer) Name() string {
	return "user-service"
}
