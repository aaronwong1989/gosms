package smgp

import (
	"fmt"

	"github.com/aaronwong1989/gosms/codec"
)

type Exit MessageHeader
type ExitResp MessageHeader

func NewExit() *Exit {
	at := &Exit{PacketLength: HeadLength, RequestId: CmdExit, SequenceId: uint32(Seq32.NextVal())}
	return at
}

func NewExitResp(seq uint32) *ExitResp {
	at := &ExitResp{PacketLength: HeadLength, RequestId: CmdExitResp, SequenceId: seq}
	return at
}

func (at *Exit) Encode() []byte {
	return (*MessageHeader)(at).Encode()
}

func (at *Exit) Decode(header codec.IHead, _ []byte) error {
	h := header.(*MessageHeader)
	at.PacketLength = h.PacketLength
	at.RequestId = h.RequestId
	at.SequenceId = h.SequenceId
	return nil
}

func (at *Exit) ToResponse(_ uint32) interface{} {
	resp := ExitResp{}
	resp.PacketLength = at.PacketLength
	resp.RequestId = CmdExitResp
	resp.SequenceId = at.SequenceId
	return &resp
}

func (at *Exit) String() string {
	return fmt.Sprintf("{ PacketLength: %d, RequestId: %s, SequenceId: %d }", at.PacketLength, CommandMap[CmdExit], at.SequenceId)
}

func (resp *ExitResp) Encode() []byte {
	return (*MessageHeader)(resp).Encode()
}

func (resp *ExitResp) Decode(header codec.IHead, _ []byte) error {
	h := header.(*MessageHeader)
	// check
	resp.PacketLength = h.PacketLength
	resp.RequestId = h.RequestId
	resp.SequenceId = h.SequenceId
	return nil
}

func (resp *ExitResp) String() string {
	return fmt.Sprintf("{ PacketLength: %d,  RequestId: %s, SequenceId: %d}", resp.PacketLength, CommandMap[CmdExitResp], resp.SequenceId)
}
