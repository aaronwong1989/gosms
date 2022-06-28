package utils

import (
	"encoding/binary"
	"testing"
)

func TestToBytes(t *testing.T) {
	os := OctetString{"1234"}
	bts := os.ToBytes(5)
	t.Logf("bts: %s, len: %d, cap: %d", bts, len(bts), cap(bts))

	bts = os.ToBytes(2)
	t.Logf("bts: %s, len: %d, cap: %d", bts, len(bts), cap(bts))
}

func TestSlice(t *testing.T) {
	var arr []byte
	t.Logf("len: %d, cap: %d", len(arr), cap(arr))

	arr = append(arr, 1)
	t.Logf("len: %d, cap: %d", len(arr), cap(arr))

	arr = append(arr, 1)
	t.Logf("len: %d, cap: %d", len(arr), cap(arr))

	frame := make([]byte, 16)
	t.Logf("len: %d, cap: %d", len(frame), cap(frame))

	binary.BigEndian.PutUint32(frame[0:4], 16)
	t.Logf("len: %d, cap: %d", len(frame), cap(frame))
	binary.BigEndian.PutUint32(frame[4:8], 1)
	t.Logf("len: %d, cap: %d", len(frame), cap(frame))
	binary.BigEndian.PutUint32(frame[8:12], 1)
	t.Logf("len: %d, cap: %d", len(frame), cap(frame))

	frame = append(frame, 101)
	t.Logf("len: %d, cap: %d", len(frame), cap(frame))

}
