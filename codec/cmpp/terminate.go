package cmpp

func NewTerminate() *MessageHeader {
	t := MessageHeader{}
	t.TotalLength = 12
	t.SequenceId = uint32(Seq32.NextVal())
	t.CommandId = CMPP_TERMINATE
	return &t
}

func NewTerminateResp(seq uint32) *MessageHeader {
	t := MessageHeader{}
	t.TotalLength = 12
	t.SequenceId = seq
	t.CommandId = CMPP_TERMINATE_RESP
	return &t
}
