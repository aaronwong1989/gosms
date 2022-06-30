package cmcc

import (
	"strconv"
	"testing"
	"time"
)

func TestCmppConnect_Encode(t *testing.T) {
	header := MessageHeader{
		TotalLength: 39,
		CommandId:   CMPP_CONNECT,
		SequenceId:  uint32(Sequence32.NextVal()),
	}

	connect := &CmppConnect{MessageHeader: &header}
	connect.sourceAddr = "123456"
	connect.version = 0x30
	connect.timestamp = uint32(1001235010)
	md5str := reqAuthMd5(connect)
	connect.authenticatorSource = string(md5str[:])
	t.Logf("%s", connect)

	frame := connect.Encode()
	t.Logf("CmppConnect: %x", frame)
}

func TestTime(t *testing.T) {
	ti := time.Now()
	// 2016-01-02 15:04:05
	s := ti.Format("0102150405")
	ts, _ := strconv.ParseUint(s, 10, 32)
	t.Logf("%T,%v", ts, ts)
	ts32 := uint32(ts)
	t.Logf("%T,%v", ts32, ts32)
}
