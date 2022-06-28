package cmcc

import (
	"encoding/binary"
	"testing"
)

func TestMessageHeader_Encode(t *testing.T) {
	header := MessageHeader{
		TotalLength: uint32(16),
		CommandId:   CMPP_CONNECT,
		SequenceId:  uint32(1),
	}
	t.Logf("%s", header.String())
	t.Logf("%d", header.Encode())

	connect := CmppConnect{MessageHeader: &header}

	connect.Encode()
}

func TestMessageHeader_Decode(t *testing.T) {
	frame := make([]byte, 16)
	binary.BigEndian.PutUint32(frame[0:4], 16)
	binary.BigEndian.PutUint32(frame[4:8], CMPP_CONNECT)
	binary.BigEndian.PutUint32(frame[8:12], 1)
	copy(frame[12:16], "1234")

	header := MessageHeader{}
	_ = header.Decode(frame)
	t.Logf("%v", header)
	t.Logf("%s", frame[12:16])
	t.Logf("% x", frame[12:16])
}
