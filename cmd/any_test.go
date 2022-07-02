package cmd

import (
	"testing"
)

func TestAnything(t *testing.T) {
	sl := []byte{0, 1, 2, 3, 4, 5, 6, 7, 8, 9}
	s1 := sl[0:5]
	s2 := sl[5:10]
	t.Logf("s1: %x", s1)
	t.Logf("s2: %x", s2)
}
