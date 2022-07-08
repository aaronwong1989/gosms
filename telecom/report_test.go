package telecom

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestReport(t *testing.T) {
	id := MsgIdSeq.NextSeq()
	rpt := NewReport(id)
	value := rpt.Encode()
	t.Logf("value: %s", value)
	rpt2 := &Report{}
	err := rpt2.Decode(value)
	assert.True(t, err == nil)
	assert.True(t, *rpt == *rpt2)
}
