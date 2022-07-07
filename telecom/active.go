package telecom

import (
	"fmt"
)

type ActiveTest MessageHeader
type ActiveTestResp MessageHeader

func NewActiveTest() *ActiveTest {
	at := &ActiveTest{PacketLength: HeadLength, RequestId: CmdActiveTest, SequenceId: uint32(Sequence32.NextVal())}
	return at
}

func (at *ActiveTest) Encode() []byte {
	return (*MessageHeader)(at).Encode()
}

func (at *ActiveTest) Decode(header *MessageHeader, _ []byte) error {
	at.PacketLength = header.PacketLength
	at.RequestId = header.RequestId
	at.SequenceId = header.SequenceId
	return nil
}

func (at *ActiveTest) ToResponse(_ uint32) interface{} {
	resp := ActiveTestResp{}
	resp.PacketLength = at.PacketLength
	resp.RequestId = CmdActiveTestResp
	resp.SequenceId = at.SequenceId
	return &resp
}

func (at *ActiveTest) String() string {
	return fmt.Sprintf("{ PacketLength: %d, RequestId: %s, SequenceId: %d }", at.PacketLength, CommandMap[CmdActiveTest], at.SequenceId)
}

func (resp *ActiveTestResp) Encode() []byte {
	return (*MessageHeader)(resp).Encode()
}

func (resp *ActiveTestResp) Decode(header *MessageHeader, _ []byte) error {
	resp.PacketLength = header.PacketLength
	resp.RequestId = header.RequestId
	resp.SequenceId = header.SequenceId
	return nil
}

func (resp *ActiveTestResp) String() string {
	return fmt.Sprintf("{ PacketLength: %d,  RequestId: %s, SequenceId: %d}", resp.PacketLength, CommandMap[CmdActiveTestResp], resp.SequenceId)
}
