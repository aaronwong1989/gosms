package snowflake32

import (
	"testing"
)

func Test_bcdToString(t *testing.T) {
	bcd := []byte{0x01, 0x23, 0x45, 0x67, 0x8a}
	str := BcdToString(bcd)

	t.Logf("bcd: %x, len: %d", bcd, len(bcd))
	t.Logf("str: %s", str)
}

func Test_stoBcd(t *testing.T) {
	str := "012345678a"
	bcd := StoBcd(str)
	t.Logf("str: %s", str)
	t.Logf("bcd: %x, len: %d", bcd, len(bcd))
}
