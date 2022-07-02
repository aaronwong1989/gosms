package cmcc

type Codec interface {
	Encode() []byte
	Decode(frame []byte) error
}
