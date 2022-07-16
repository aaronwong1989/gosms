package smgp

import (
	"testing"
	"time"
)

func TestMtOptions(t *testing.T) {
	var ti time.Time
	t.Logf("%v", ti)
	t.Logf("%v", ti.Year())

	var d time.Duration
	t.Logf("%v", d)
}
