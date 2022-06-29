package cmcc

import (
	"fmt"
)

type ActiveTest struct {
	*MessageHeader
}

func NewActiveTest() *ActiveTest {
	header := &MessageHeader{TotalLength: HEAD_LENGTH, CommandId: CMPP_ACTIVE_TEST, SequenceId: uint32(Sequence.NextVal())}
	return &ActiveTest{header}
}

func (at *ActiveTest) Encode() []byte {
	return at.MessageHeader.Encode()
}

func (at *ActiveTest) Decode(frame []byte) error {
	return at.MessageHeader.Decode(frame)
}

func (at *ActiveTest) ToResponse() *ActiveTestResp {
	header := &MessageHeader{TotalLength: HEAD_LENGTH + 1, CommandId: CMPP_ACTIVE_TEST_RESP, SequenceId: at.SequenceId}
	return &ActiveTestResp{MessageHeader: header, Reserved: 0}
}

func (at *ActiveTest) String() string {
	return fmt.Sprintf("{ TotalLength: %d, CommandId: CMPP_ACTIVE_TEST, SequenceId: %d }", at.TotalLength, at.SequenceId)
}

type ActiveTestResp struct {
	*MessageHeader
	Reserved byte
}

func (at *ActiveTestResp) Encode() []byte {
	return at.MessageHeader.Encode()
}

func (at *ActiveTestResp) Decode(frame []byte) error {
	return at.MessageHeader.Decode(frame)
}

func (at *ActiveTestResp) String() string {
	return fmt.Sprintf("{ TotalLength: %d, CommandId: CMPP_ACTIVE_TEST_RESP, SequenceId: %d, Reserved: %d }", at.TotalLength, at.SequenceId, at.Reserved)
}
