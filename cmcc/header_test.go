package cmcc

import (
	"encoding/binary"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMessageHeader_Encode(t *testing.T) {
	header := MessageHeader{
		TotalLength: 16,
		CommandId:   CMPP_CONNECT,
		SequenceId:  uint32(Sequence32.NextVal()),
	}
	t.Logf("%s", header.String())
	t.Logf("%d", header.Encode())

	connect := Connect{MessageHeader: &header}

	connect.Encode()

	assert.Equal(t, int(Conf.Version)&0xf0, 0x20)

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

func TestStringRune(t *testing.T) {
	str := "中国人hello"
	bts := []byte(str)
	chars := []rune(str)

	t.Logf("len(str)=%d,len(bts)=%d,len(chars)=%d", len(str), len(bts), len(chars))

	t.Logf("bts: %x", bts)
	t.Logf("chars: %c", chars)
}

func TestTrimStr(t *testing.T) {
	bts := []byte{'a', 'b', 'c', 'd', 0, 0, 0}
	t.Logf("%s", TrimStr(bts))
}
