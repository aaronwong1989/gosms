package telecom

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDeliver_Decode(t *testing.T) {
	dlv := NewDeliver("123", "95535", "TD:123456")
	t.Logf("dlv: %s", dlv)
	testDeliver(t, dlv)
}

func TestDeliver_ReportDecode(t *testing.T) {
	mts := NewSubmit([]string{"17011113333"}, "hello world，世界", MtOptions{})
	mt := mts[0]
	msp := mt.ToResponse(0).(*SubmitResp)
	rpt := NewDeliveryReport(mt, msp.msgId)
	t.Logf("dlv: %s", rpt)
	testDeliver(t, rpt)
}

func testDeliver(t *testing.T, dlv *Deliver) {
	resp := dlv.ToResponse(0).(*DeliverResp)
	t.Logf("resp: %s", resp)

	// 测试Deliver Encode
	dt := dlv.Encode()
	assert.True(t, int(dlv.PacketLength) == len(dt))
	t.Logf("dlv_encode: %x", dt)
	// 测试Deliver Decode
	h := &MessageHeader{}
	err := h.Decode(dt)
	assert.True(t, err == nil)
	dlvDec := &Deliver{}
	err = dlvDec.Decode(h, dt[12:])
	assert.True(t, err == nil)
	assert.True(t, dlvDec.MessageHeader.SequenceId == dlv.MessageHeader.SequenceId)
	t.Logf("dlv_decode: %s", dlvDec)

	// 测试DeliverResp Encode
	dt = resp.Encode()
	assert.True(t, int(resp.PacketLength) == len(dt))
	t.Logf("resp_encode: %x", dt)
	// 测试Deliver Decode
	h = &MessageHeader{}
	err = h.Decode(dt)
	assert.True(t, err == nil)
	respDec := &DeliverResp{}
	err = respDec.Decode(h, dt[12:])
	assert.True(t, err == nil)
	assert.True(t, respDec.MessageHeader.SequenceId == respDec.MessageHeader.SequenceId)
	t.Logf("resp_decode: %s", dlvDec)
}
