package smgp

import (
	"fmt"
)

type Codec interface {
	Encode() []byte
	Decode(header *MessageHeader, frame []byte) error
}

type Pdu interface {
	Codec
	fmt.Stringer
	ToResponse(code uint32) interface{}
}

type Sequence32 interface {
	NextVal() int32
}

type Sequence80 interface {
	NextVal() []byte
}
