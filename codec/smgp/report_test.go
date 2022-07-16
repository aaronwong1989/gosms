package smgp

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestReport(t *testing.T) {
	id := Seq80.NextVal()
	rpt := NewReport(id)
	t.Logf("rpt: %s", rpt)
	data := rpt.Encode()
	assert.True(t, len(data) == RptLen)
	t.Logf("value: %x", data)

	rpt2 := &Report{}
	err := rpt2.Decode(data)
	assert.True(t, err == nil)
	t.Logf("rpt2: %s", rpt2)
}
