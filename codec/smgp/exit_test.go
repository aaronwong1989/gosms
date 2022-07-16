package smgp

import (
	"testing"
)

func TestExit(t *testing.T) {
	exit := NewExit()
	t.Logf("%T : %s", exit, exit)

	data := exit.Encode()
	t.Logf("%T : %x", data, data)

	h := &MessageHeader{}
	_ = h.Decode(data)
	t.Logf("%T : %s", h, h)

	e2 := &Exit{}
	_ = e2.Decode(h, data)
	t.Logf("%T : %s", e2, e2)

	resp := exit.ToResponse(0).(*ExitResp)
	t.Logf("%T : %s", resp, resp)

	data = resp.Encode()
	t.Logf("%T : %x", data, data)
	_ = h.Decode(data)

	resp2 := &ExitResp{}
	_ = resp2.Decode(h, data)
	t.Logf("%T : %s", resp2, resp2)
}
