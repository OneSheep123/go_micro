package rpc

import "context"

type onewayKey struct {
}

func CtxWithOneWay(ctx context.Context) context.Context {
	// 推荐：使用结构体作为key
	return context.WithValue(ctx, onewayKey{}, true)
}

func isOneWay(ctx context.Context) bool {
	val := ctx.Value(onewayKey{})
	oneway, ok := val.(bool)
	return ok && oneway
}
