package comm

import (
	"fmt"
	"testing"
)

func Test_BcdToString(t *testing.T) {
	bcd := []byte{0x01, 0x23, 0x45, 0x67, 0x8a}
	str := BcdToString(bcd)

	t.Logf("bcd: %x, len: %d", bcd, len(bcd))
	t.Logf("str: %s", str)
}

func Test_StoBcd(t *testing.T) {
	str := "012345678a"
	bcd := StoBcd(str)
	t.Logf("str: %s", str)
	t.Logf("bcd: %x, len: %d", bcd, len(bcd))
}

var bcdSeq = NewBcdSequence("000001")

func BenchmarkBcdSequence_NextSeq(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		bcdSeq.NextSeq()
	}
}

func BenchmarkBcdSequence_BcdToString(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		BcdToString(bcdSeq.NextSeq())
	}
}

func BenchmarkBcdSequence_FmtString(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = fmt.Sprintf("%x", bcdSeq.NextSeq())
	}
}

func BenchmarkBcdSequence_BcdToStringParallel(b *testing.B) {
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			BcdToString(bcdSeq.NextSeq())
		}
	})
}

func BenchmarkBcdSequence_FmtStringParallel(b *testing.B) {
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_ = fmt.Sprintf("%x", bcdSeq.NextSeq())
		}
	})
}
