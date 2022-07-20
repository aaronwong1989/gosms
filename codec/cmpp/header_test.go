package cmpp

import (
	"encoding/binary"
	"math/rand"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/aaronwong1989/gosms/comm"
	"github.com/aaronwong1989/gosms/comm/snowflake"
	"github.com/aaronwong1989/gosms/comm/yml_config"
)

func init() {
	rand.Seed(time.Now().Unix()) // 随机种子
	Conf = yml_config.CreateYamlFactory("cmpp.yaml")
	dc := Conf.GetInt("data-center-id")
	wk := Conf.GetInt("worker-id")
	Seq32 = comm.NewCycleSequence(int32(dc), int32(wk))
	Seq64 = snowflake.NewSnowflake(int64(dc), int64(wk))
	ReportSeq = comm.NewCycleSequence(int32(dc), int32(wk))
}

func TestMessageHeader_Encode(t *testing.T) {
	header := MessageHeader{
		TotalLength: 16,
		CommandId:   CMPP_CONNECT,
		SequenceId:  uint32(Seq32.NextVal()),
	}
	t.Logf("%s", header.String())
	t.Logf("%d", header.Encode())

	connect := Connect{MessageHeader: &header}

	connect.Encode()

	assert.Equal(t, int(Conf.GetInt("version"))&0xf0, 0x20)

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
