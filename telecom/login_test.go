package telecom

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLogin_Decode(t *testing.T) {
	var i = 100
	for i > 0 {
		lo := NewLogin()
		t.Logf("login   : %s", lo)
		assert.True(t, lo.clientID == Conf.ClientId)
		resp := lo.ToResponse(0).(*LoginResp)
		t.Logf("resp    : %s", resp)
		assert.True(t, lo.clientID == Conf.ClientId)

		dt1 := lo.Encode()
		dt2 := resp.Encode()
		assert.True(t, len(dt1) == LoginLen)
		assert.True(t, len(dt2) == LoginRespLen)

		err := lo.Decode(lo.MessageHeader, dt1[12:])
		assert.True(t, err == nil)
		t.Logf("loginDec: %s, err: %s", lo, err)
		err = resp.Decode(resp.MessageHeader, dt2[12:])
		assert.True(t, err == nil)
		t.Logf("respDec : %s, err: %s", resp, err)
		i--
	}
}
