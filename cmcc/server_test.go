package main

import (
	"net"
	"testing"

	"github.com/stretchr/testify/assert"

	cmcc "sms-adapter/cmcc/protocol"
)

func TestSendConnect(t *testing.T) {
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
	con := cmcc.NewConnect()
	t.Logf("send: %s", con)
	i, _ := c.Write(con.Encode())
	assert.True(t, uint32(i) == con.TotalLength)
	resp := make([]byte, 33)
	i, _ = c.Read(resp)
	assert.True(t, i == 33)

	header := &cmcc.MessageHeader{}
	err = header.Decode(resp)
	if err != nil {
		return
	}
	pdu := &cmcc.CmppConnectResp{}
	err = pdu.Decode(header, resp[12:])
	if err != nil {
		return
	}
	t.Logf("receive: %s", pdu)
}
