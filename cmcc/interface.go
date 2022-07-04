package cmcc

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
