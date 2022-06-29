package main

import (
	"net"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	cmcc "sms-vgateway/cmcc/protocol"
)

func TestClient(t *testing.T) {
	concurrency := 4
	var wg sync.WaitGroup
	wg.Add(concurrency)
	for i := 0; i < concurrency; i++ {
		runClient(t, &wg)
		time.Sleep(time.Millisecond * 10)
	}
	wg.Wait()
}

func runClient(t *testing.T, wg *sync.WaitGroup) {
	go func(t *testing.T) {
		c, err := net.Dial("tcp", ":9000")
		if err != nil {
			log.Errorf("%v", err)
		}
		defer func(c net.Conn) {
			err := c.Close()
			if err != nil {
				log.Errorf("%v", err)
			}
		}(c)

		sendConnect(t, c)

		for {
			bytes := make([]byte, 12)
			_, err := c.Read(bytes)
			if err != nil {
				return
			}
			t.Logf("read %x from %s", bytes, c.RemoteAddr().String())
			header := &cmcc.MessageHeader{}
			_ = header.Decode(bytes)
			if header.CommandId == cmcc.CMPP_ACTIVE_TEST {
				at := cmcc.ActiveTest{MessageHeader: header}
				_, _ = c.Write(at.ToResponse().Encode())
			}
		}
		wg.Done()
	}(t)
}

func sendConnect(t *testing.T, c net.Conn) {
	con := cmcc.NewConnect()
	con.AuthenticatorSource = "000000" // 反例测试
	t.Logf("send: %s", con)
	i, _ := c.Write(con.Encode())
	assert.True(t, uint32(i) == con.TotalLength)
	resp := make([]byte, cmcc.LEN_CMPP_CONNECT_RESP)
	i, _ = c.Read(resp)
	assert.True(t, i == cmcc.LEN_CMPP_CONNECT_RESP)

	header := &cmcc.MessageHeader{}
	err := header.Decode(resp)
	if err != nil {
		return
	}
	rep := &cmcc.CmppConnectResp{}
	err = rep.Decode(header, resp[cmcc.HEAD_LENGTH:])
	if err != nil {
		return
	}
	t.Logf("receive: %s", rep)
	assert.True(t, 0 == rep.Status)
}
