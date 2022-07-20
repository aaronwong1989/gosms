package cmpp

import (
	"fmt"

	"github.com/aaronwong1989/gosms/codec"
)

type ActiveTest struct {
	*MessageHeader
}

func NewActiveTest() *ActiveTest {
	header := &MessageHeader{TotalLength: HeadLength, CommandId: CMPP_ACTIVE_TEST, SequenceId: uint32(Seq32.NextVal())}
	return &ActiveTest{header}
}

func (at *ActiveTest) Encode() []byte {
	return at.MessageHeader.Encode()
}

func (at *ActiveTest) Decode(header codec.IHead, frame []byte) error {
	h := header.(*MessageHeader)
	if header == nil || h.CommandId != CMPP_ACTIVE_TEST || frame != nil {
		return ErrorPacket
	}
	at.MessageHeader = h
	return nil
}

func (at *ActiveTest) ToResponse(_ uint32) interface{} {
	header := &MessageHeader{TotalLength: HeadLength + 1, CommandId: CMPP_ACTIVE_TEST_RESP, SequenceId: at.SequenceId}
	atr := ActiveTestResp{MessageHeader: header, reserved: 0}
	return &atr
}

func (at *ActiveTest) String() string {
	return fmt.Sprintf("{ TotalLength: %d, CommandId: CMPP_ACTIVE_TEST, SequenceId: %d }", at.TotalLength, at.SequenceId)
}

type ActiveTestResp struct {
	*MessageHeader
	reserved byte
}

func (at *ActiveTestResp) Encode() []byte {
	return at.MessageHeader.Encode()
}

func (at *ActiveTestResp) Decode(header codec.IHead, frame []byte) error {
	h := header.(*MessageHeader)
	if header == nil || h.CommandId != CMPP_ACTIVE_TEST_RESP || len(frame) < (13-HeadLength) {
		return ErrorPacket
	}
	at.MessageHeader = h
	at.reserved = frame[0]
	return nil
}

func (at *ActiveTestResp) String() string {
	return fmt.Sprintf("{ TotalLength: %d, CommandId: CMPP_ACTIVE_TEST_RESP, SequenceId: %d, reserved: %d }", at.TotalLength, at.SequenceId, at.reserved)
}
