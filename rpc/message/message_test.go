// create by chencanhua in 2024/5/2
package message

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestEncodeDecodeRequest(t *testing.T) {
	testCases := []struct {
		name string
		req  *Request
	}{
		{
			name: "with meta",
			req: &Request{
				MessageId:   123,
				Version:     12,
				Compresser:  25,
				Serializer:  17,
				ServiceName: "user-service",
				MethodName:  "GetById",
				Meta: map[string]string{
					"trace-id": "123",
					"a/b":      "b",
					"shadow":   "true",
				},
				Data: []byte("hello, world"),
			},
		},
		{
			name: "no meta",
			req: &Request{
				MessageId:   123,
				Version:     12,
				Compresser:  25,
				Serializer:  17,
				ServiceName: "user-service",
				MethodName:  "GetById",
				Data:        []byte("hello, world"),
			},
		},
		{
			name: "empty value",
			req: &Request{
				MessageId:   123,
				Version:     12,
				Compresser:  25,
				Serializer:  17,
				ServiceName: "user-service",
				MethodName:  "GetById",
				Meta: map[string]string{
					"trace-id": "123",
					"a/b":      "",
					"shadow":   "true",
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.req.SetHeadLength()
			tc.req.SetBodyLength()
			bs := EncodeReq(tc.req)
			req := DecodeReq(bs)
			assert.Equal(t, tc.req, req)
		})
	}
}

func TestEncodeDecodeResponse(t *testing.T) {
	testCases := []struct {
		name string
		resp *Response
	}{
		{
			name: "with no error",
			resp: &Response{
				MessageId:  123,
				Version:    12,
				Compresser: 25,
				Serializer: 17,
				Data:       []byte("hello, world"),
			},
		},
		{
			name: "error",
			resp: &Response{
				MessageId:  123,
				Version:    12,
				Compresser: 25,
				Serializer: 17,
				Error:      []byte("123"),
			},
		},
		{
			name: "error and data",
			resp: &Response{
				MessageId:  123,
				Version:    12,
				Compresser: 25,
				Serializer: 17,
				Error:      []byte("123"),
				Data:       []byte("hello, world"),
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.resp.SetHeadLength()
			tc.resp.SetBodyLength()
			bs := EncodeResp(tc.resp)
			resp := DecodeResp(bs)
			assert.Equal(t, tc.resp, resp)
		})
	}
}
