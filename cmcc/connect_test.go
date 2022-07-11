package cmcc

import (
	"crypto/md5"
	"fmt"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestCmppConnect_Encode(t *testing.T) {
	header := MessageHeader{
		TotalLength: 39,
		CommandId:   CMPP_CONNECT,
		SequenceId:  uint32(RequestSeq.NextVal()),
	}

	connect := &Connect{MessageHeader: &header}
	connect.sourceAddr = "123456"
	connect.version = 0x20
	connect.timestamp = uint32(1001235010)
	md5str := reqAuthMd5(connect)
	connect.authenticatorSource = md5str[:]
	t.Logf("%s", connect)

	frame := connect.Encode()
	t.Logf("Connect: %x", frame)
	assert.Equal(t, uint32(0), connect.Check())

	resp := connect.ToResponse(0).(*ConnectResp)
	t.Logf("Connect: %v", resp)
	t.Logf("ConnectResp: %x", resp.Encode())
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

func TestAuthStr(t *testing.T) {
	ti := time.Now()
	s := ti.Format("0102150405")
	ts, _ := strconv.ParseUint(s, 10, 32)
	t.Logf("%T, %v, %s", ts, ts, fmt.Sprintf("%010d", ts))

	authDt := make([]byte, 0, 64)
	authDt = append(authDt, "901234"...)
	authDt = append(authDt, 0, 0, 0, 0, 0, 0, 0, 0, 0)
	authDt = append(authDt, "1234"...)
	authDt = append(authDt, "0706104024"...)
	authMd5 := md5.Sum(authDt)

	t.Logf("%x", authMd5)
}
