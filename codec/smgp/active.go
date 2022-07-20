package smgp

import (
	"fmt"

	"github.com/aaronwong1989/gosms/codec"
)

type ActiveTest MessageHeader
type ActiveTestResp MessageHeader

func NewActiveTest() *ActiveTest {
	at := &ActiveTest{PacketLength: HeadLength, RequestId: CmdActiveTest, SequenceId: uint32(Seq32.NextVal())}
	return at
}

func NewActiveTestResp(seq uint32) *ActiveTestResp {
	at := &ActiveTestResp{PacketLength: HeadLength, RequestId: CmdActiveTest, SequenceId: seq}
	return at
}

func (at *ActiveTest) Encode() []byte {
	return (*MessageHeader)(at).Encode()
}

func (at *ActiveTest) Decode(header codec.IHead, _ []byte) error {
	h := header.(*MessageHeader)
	at.PacketLength = h.PacketLength
	at.RequestId = h.RequestId
	at.SequenceId = h.SequenceId
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

func (resp *ActiveTestResp) Decode(header codec.IHead, _ []byte) error {
	h := header.(*MessageHeader)
	resp.PacketLength = h.PacketLength
	resp.RequestId = h.RequestId
	resp.SequenceId = h.SequenceId
	return nil
}

func (resp *ActiveTestResp) String() string {
	return fmt.Sprintf("{ PacketLength: %d,  RequestId: %s, SequenceId: %d}", resp.PacketLength, CommandMap[CmdActiveTestResp], resp.SequenceId)
}
