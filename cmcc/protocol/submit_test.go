package cmcc

import (
	"reflect"
	"testing"
)

func TestEncode(t *testing.T) {
	type args struct {
		s string
	}
	tests := []struct {
		name string
		args args
		want []byte
	}{
		{
			"testcase1",
			args{"1"},
			[]byte{0x00, 0x31},
		},
		{
			"testcase2",
			args{"hello world"},
			[]byte{0x00, 0x68, 0x00, 0x65, 0x00, 0x6c, 0x00, 0x6c, 0x00, 0x6f, 0x00, 0x20, 0x00, 0x77, 0x00, 0x6f, 0x00, 0x72, 0x00, 0x6c, 0x00, 0x64},
		},
		{"testcase3",
			args{"Great 中国"},
			[]byte{0x00, 0x47, 0x00, 0x72, 0x00, 0x65, 0x00, 0x61, 0x00, 0x74, 0x00, 0x20, 0x4e, 0x2d, 0x56, 0xfd},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ucs2Encode(tt.args.s); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ucs2Encode() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDecode(t *testing.T) {
	type args struct {
		ucs2 []byte
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			"testcase1",
			args{[]byte{0x00, 0x31}},
			"1",
		},
		{
			"testcase2",
			args{[]byte{0x00, 0x68, 0x00, 0x65, 0x00, 0x6c, 0x00, 0x6c, 0x00, 0x6f, 0x00, 0x20, 0x00, 0x77, 0x00, 0x6f, 0x00, 0x72, 0x00, 0x6c, 0x00, 0x64}},
			"hello world",
		},
		{"testcase3",
			args{[]byte{0x00, 0x47, 0x00, 0x72, 0x00, 0x65, 0x00, 0x61, 0x00, 0x74, 0x00, 0x20, 0x4e, 0x2d, 0x56, 0xfd}},
			"Great 中国",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ucs2Decode(tt.args.ucs2); got != tt.want {
				t.Errorf("ucs2Decode() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNewSubmit(t *testing.T) {
	phones := []string{"17011112222", "17500002222"}
	// 160 bytes
	content := "0123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789"
	mts := NewSubmit(phones, content)
	for _, mt := range mts {
		t.Logf(">>> %s", mt)
		t.Logf("<<< %s", mt.ToResponse(0))
	}

	content2 := content + "hello world"
	mts = NewSubmit(phones, content2)
	for _, mt := range mts {
		t.Logf(">>> %s", mt)
		t.Logf("<<< %s", mt.ToResponse(0))
	}

	content3 := "强大的祖国"
	mts = NewSubmit(phones, content3)
	for _, mt := range mts {
		t.Logf(">>> %s", mt)
		t.Logf("<<< %s", mt.ToResponse(0))
	}

	content4 := "强大的祖国" + content
	mts = NewSubmit(phones, content4)
	for _, mt := range mts {
		t.Logf(">>> %s", mt)
		t.Logf("<<< %s", mt.ToResponse(0))
	}
}
