package net

import (
	"context"
	"errors"
	"net"
	"sync"
	"time"
)

type Pool struct {
	// 空闲连接队列
	idlesConns chan *idleConn

	// 请求队列
	reqQueue []*connReq

	// 最大连接数
	maxCnt int

	// 当前连接数
	cnt int

	// 最大空闲时间
	maxIdleTime time.Duration

	factory func() (net.Conn, error)

	lock *sync.Mutex
}

func NewPool(initCnt int, maxIdleCnt int, maxCnt int,
	maxIdleTime time.Duration,
	factory func() (net.Conn, error)) (*Pool, error) {
	if initCnt > maxCnt {
		return nil, errors.New("micro: 初始连接数量不能超过最大连接数")
	}
	if initCnt > maxIdleCnt {
		return nil, errors.New("micro: 初始连接数量不能大于最大空闲连接数量")
	}

	idlesConns := make(chan *idleConn, maxIdleCnt)

	for i := 0; i < initCnt; i++ {
		conn, err := factory()
		if err != nil {
			return nil, err
		}
		idlesConns <- &idleConn{c: conn, lastActiveTime: time.Now()}
	}

	res := &Pool{
		idlesConns:  idlesConns,
		maxCnt:      maxCnt,
		cnt:         0,
		maxIdleTime: maxIdleTime,
		factory:     factory,
		lock:        &sync.Mutex{},
	}
	return res, nil
}

func (p *Pool) Get(ctx context.Context) (net.Conn, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	for {
		select {
		case conn := <-p.idlesConns:
			// 当前连接已经无效
			if conn.lastActiveTime.Add(p.maxIdleTime).Before(time.Now()) {
				_ = conn.c.Close()
				continue
			}
			return conn.c, nil
		default:
			p.lock.Lock()
			// 当前连接已经超过最大连接数
			if p.cnt >= p.maxCnt {
				req := connReq{connChan: make(chan net.Conn, 1)}
				p.reqQueue = append(p.reqQueue, &req)
				//
				p.lock.Unlock()
				select {
				// 等待别人把连接赋值到当前请求连接中
				case c := <-req.connChan:
					return c, nil
				// 等待超时了
				case <-ctx.Done():
					go func() {
						// 并发原因，即使超时了也有可能有人给连接给你
						c := <-req.connChan
						p.Put(context.Background(), c)
					}()
				}
			}
			// 进行新建连接
			conn, err := p.factory()
			if err != nil {
				return nil, err
			}
			p.cnt++
			p.lock.Unlock()
			return conn, nil
		}
	}
}

func (p *Pool) Put(ctx context.Context, conn net.Conn) error {
	p.lock.Lock()
	if len(p.reqQueue) > 0 {
		req := p.reqQueue[len(p.reqQueue)-1]
		p.reqQueue = p.reqQueue[:len(p.reqQueue)-1]
		// 注意：先进行解锁
		p.lock.Unlock()
		req.connChan <- conn
		return nil
	}

	defer p.lock.Unlock()
	ic := &idleConn{
		c:              conn,
		lastActiveTime: time.Now(),
	}
	select {
	case p.idlesConns <- ic:
	default:
		_ = conn.Close()
		p.cnt--
	}
	return nil
}

// idleConn 连接
type idleConn struct {
	c              net.Conn
	lastActiveTime time.Time
}

// connReq 连接请求
type connReq struct {
	connChan chan net.Conn
}
