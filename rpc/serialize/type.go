// create by chencanhua in 2024/5/3
package serialize

type Serialize interface {
	Code() byte
	Encode(val any) ([]byte, error)
	Decode(data []byte, val any) error
}
