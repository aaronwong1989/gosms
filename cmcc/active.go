package cmcc

import (
	"fmt"
)

type ActiveTest struct {
	*MessageHeader
}

func NewActiveTest() *ActiveTest {
	header := &MessageHeader{TotalLength: HEAD_LENGTH, CommandId: CMPP_ACTIVE_TEST, SequenceId: uint32(RequestSeq.NextVal())}
	return &ActiveTest{header}
}

func (at *ActiveTest) Encode() []byte {
	return at.MessageHeader.Encode()
}

func (at *ActiveTest) Decode(header *MessageHeader, frame []byte) error {
	if header == nil || header.CommandId != CMPP_ACTIVE_TEST || frame != nil {
		return ErrorPacket
	}
	at.MessageHeader = header
	return nil
}

func (at *ActiveTest) ToResponse(_ uint32) interface{} {
	header := &MessageHeader{TotalLength: HEAD_LENGTH + 1, CommandId: CMPP_ACTIVE_TEST_RESP, SequenceId: at.SequenceId}
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

func (at *ActiveTestResp) Decode(header *MessageHeader, frame []byte) error {
	if header == nil || header.CommandId != CMPP_ACTIVE_TEST_RESP || len(frame) < (13-HEAD_LENGTH) {
		return ErrorPacket
	}
	at.MessageHeader = header
	at.reserved = frame[0]
	return nil
}

func (at *ActiveTestResp) String() string {
	return fmt.Sprintf("{ TotalLength: %d, CommandId: CMPP_ACTIVE_TEST_RESP, SequenceId: %d, reserved: %d }", at.TotalLength, at.SequenceId, at.reserved)
}
