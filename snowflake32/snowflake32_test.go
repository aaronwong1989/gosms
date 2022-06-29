package snowflake32

import (
	"testing"
	"time"
)

func TestTimeStamp(t *testing.T) {
	snow := &Snowflake{datacenter: 1, worker: 1}
	time.Sleep(time.Second)
	for i := 0; i < 520; i++ {
		t.Logf("curVal: %#b", snow.NextVal())
		t.Logf("curVal: %s", snow)
	}
	seq := uint32(snow.NextVal())
	t.Logf("uint32 seq: %d", seq)
	t.Logf("uint32 seq: %d", snow.NextVal())
}
