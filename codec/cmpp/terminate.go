package cmpp

import (
	"github.com/aaronwong1989/gosms/codec"
)

func NewTerminate() codec.IHead {
	t := MessageHeader{}
	t.TotalLength = 12
	t.SequenceId = uint32(Seq32.NextVal())
	t.CommandId = CMPP_TERMINATE
	return &t
}

func NewTerminateResp(seq uint32) codec.IHead {
	t := MessageHeader{}
	t.TotalLength = 12
	t.SequenceId = seq
	t.CommandId = CMPP_TERMINATE_RESP
	return &t
}
