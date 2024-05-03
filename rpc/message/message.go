package message

import (
	"bytes"
	"encoding/binary"
)

// 头部不定长字段的分隔符
const (
	splitter     = '\n'
	pairSplitter = '\r'
)

type Request struct {
	// 头部
	// 消息长度
	HeadLength uint32
	// 协议版本
	BodyLength uint32
	// 消息ID
	MessageId uint32
	// 版本, 一个字节
	Version uint8
	// 压缩算法
	Compresser uint8
	// 序列化方法
	Serializer uint8

	// 服务名称和方法名称
	ServiceName string
	MethodName  string

	// 扩展字段，用于传毒自定义元数据(key:value)
	Meta map[string]string

	// 数据部
	Data []byte
}

func (req *Request) SetHeadLength() {
	// uint32 => 4个字节
	res := 15
	res += len(req.ServiceName)
	// 分隔符
	res++
	res += len(req.MethodName)
	// 分隔符
	res++
	for key, value := range req.Meta {
		res += len(key)
		res++
		res += len(value)
		res++
	}
	req.HeadLength = uint32(res)
}

func (req *Request) SetBodyLength() {
	req.BodyLength = uint32(len(req.Data))
}

func EncodeReq(req *Request) []byte {
	bs := make([]byte, req.HeadLength+req.BodyLength)
	cur := bs

	binary.BigEndian.PutUint32(cur[:4], req.HeadLength)
	binary.BigEndian.PutUint32(cur[4:8], req.BodyLength)
	binary.BigEndian.PutUint32(cur[8:12], req.MessageId)
	cur[12] = req.Version
	cur[13] = req.Compresser
	cur[14] = req.Serializer
	cur = cur[15:]

	copy(cur, req.ServiceName)
	cur[len(req.ServiceName)] = splitter
	cur = cur[len(req.ServiceName)+1:]

	copy(cur, req.MethodName)
	cur[len(req.MethodName)] = splitter
	cur = cur[len(req.MethodName)+1:]

	for key, value := range req.Meta {
		copy(cur, key)
		cur[len(key)] = pairSplitter
		cur = cur[len(key)+1:]

		copy(cur, value)
		cur[len(value)] = splitter
		cur = cur[len(value)+1:]
	}
	if req.BodyLength > 0 {
		// 剩下的数据
		copy(cur, req.Data)
	}
	return bs
}

func DecodeReq(data []byte) *Request {
	req := &Request{}
	req.HeadLength = binary.BigEndian.Uint32(data[:4])
	req.HeadLength = binary.BigEndian.Uint32(data[:4])
	req.BodyLength = binary.BigEndian.Uint32(data[4:8])
	req.MessageId = binary.BigEndian.Uint32(data[8:12])
	req.Version = data[12]
	req.Compresser = data[13]
	req.Serializer = data[14]
	meta := data[15:req.HeadLength]

	index := bytes.IndexByte(meta, splitter)
	req.ServiceName = string(meta[:index])
	meta = meta[index+1:]

	index = bytes.IndexByte(meta, splitter)
	req.MethodName = string(meta[:index])
	meta = meta[index+1:]

	// 继续拆解 meta 剩下的 key value
	if len(meta) > 0 {
		// 这个地方不好预估容量，但是大部分都很少，我们把现在能够想到的元数据都算法
		// 也就不超过四个
		metaMap := make(map[string]string, 4)
		// 第一对键值对出现的下标
		index = bytes.IndexByte(meta, splitter)
		for index != -1 {
			pair := meta[:index]
			pairIndex := bytes.IndexByte(pair, pairSplitter)
			metaMap[string(pair[:pairIndex])] = string(pair[pairIndex+1:])

			meta = meta[index+1:]
			index = bytes.IndexByte(meta, splitter)
		}
		req.Meta = metaMap
	}

	if req.BodyLength > 0 {
		// 剩下的就是数据了
		req.Data = data[req.HeadLength:]
	}
	return req
}

type Response struct {

	// 消息长度
	HeadLength uint32
	// 协议版本
	BodyLength uint32
	// 消息ID
	MessageId uint32
	// 版本, 一个字节
	Version uint8
	// 压缩算法
	Compresser uint8
	// 序列化方法
	Serializer uint8

	Error []byte

	Data []byte
}

func EncodeResp(resp *Response) []byte {
	bs := make([]byte, resp.HeadLength+resp.BodyLength)
	cur := bs

	binary.BigEndian.PutUint32(cur[:4], resp.HeadLength)
	binary.BigEndian.PutUint32(cur[4:8], resp.BodyLength)
	binary.BigEndian.PutUint32(cur[8:12], resp.MessageId)
	cur[12] = resp.Version
	cur[13] = resp.Compresser
	cur[14] = resp.Serializer
	cur = cur[15:]

	if resp.HeadLength > 15 {
		copy(cur, resp.Error)
		cur = cur[len(resp.Error):]
	}

	if resp.BodyLength > 0 {
		// 剩下的数据
		copy(cur, resp.Data)
	}

	return bs
}

func DecodeResp(data []byte) *Response {
	resp := &Response{}
	resp.HeadLength = binary.BigEndian.Uint32(data[:4])
	resp.HeadLength = binary.BigEndian.Uint32(data[:4])
	resp.BodyLength = binary.BigEndian.Uint32(data[4:8])
	resp.MessageId = binary.BigEndian.Uint32(data[8:12])
	resp.Version = data[12]
	resp.Compresser = data[13]
	resp.Serializer = data[14]

	if resp.HeadLength > 15 {
		resp.Error = data[15:resp.HeadLength]
	}

	if resp.BodyLength > 0 {
		// 剩下的就是数据了
		resp.Data = data[resp.HeadLength:]
	}
	return resp
}

func (r *Response) SetHeadLength() {
	// uint32 => 4个字节
	res := 15
	res += len(r.Error)
	r.HeadLength = uint32(res)
}

func (r *Response) SetBodyLength() {
	r.BodyLength = uint32(len(r.Data))
}
