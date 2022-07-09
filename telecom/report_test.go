package telecom

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestReport(t *testing.T) {
	id := MsgIdSeq.NextSeq()
	rpt := NewReport(fmt.Sprintf("%x", id))
	value := rpt.Encode()
	t.Logf("len:%d, value: %s", len(value), value)
	rpt2 := &Report{}
	err := rpt2.Decode(value)
	assert.True(t, err == nil)
	assert.True(t, *rpt == *rpt2)
}
