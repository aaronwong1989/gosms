package telecom

import (
	"fmt"
)

type Exit MessageHeader
type ExitResp MessageHeader

func NewExit() *Exit {
	at := &Exit{PacketLength: HeadLength, RequestId: CmdExit, SequenceId: uint32(RequestSeq.NextVal())}
	return at
}

func NewExitResp(seq uint32) *ExitResp {
	at := &ExitResp{PacketLength: HeadLength, RequestId: CmdExitResp, SequenceId: seq}
	return at
}

func (at *Exit) Encode() []byte {
	return (*MessageHeader)(at).Encode()
}

func (at *Exit) Decode(header *MessageHeader, _ []byte) error {
	at.PacketLength = header.PacketLength
	at.RequestId = header.RequestId
	at.SequenceId = header.SequenceId
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

func (resp *ExitResp) Decode(header *MessageHeader, _ []byte) error {
	resp.PacketLength = header.PacketLength
	resp.RequestId = header.RequestId
	resp.SequenceId = header.SequenceId
	return nil
}

func (resp *ExitResp) String() string {
	return fmt.Sprintf("{ PacketLength: %d,  RequestId: %s, SequenceId: %d}", resp.PacketLength, CommandMap[CmdExitResp], resp.SequenceId)
}
