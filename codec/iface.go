package codec

type IHead interface {
	Encode() []byte
	Decode([]byte) error
	String() string
}

type Codec interface {
	Encode() []byte
	Decode(header IHead, frame []byte) error
	String() string
}

type RequestPdu interface {
	Codec
	ToResponse(code uint32) interface{}
}

// Sequence32 32位序号生成器
type Sequence32 interface {
	NextVal() int32
}

// Sequence64 64位序号生成器
type Sequence64 interface {
	NextVal() int64
}

// SequenceBCD BCD码序号生成器
type SequenceBCD interface {
	NextVal() []byte
}
