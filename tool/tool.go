package tool

import (
	"time"
	"unsafe"
)

func TrimStr(bts []byte) string {
	var i = 0
	for ; i < len(bts); i++ {
		if bts[i] == 0 {
			break
		}
	}
	ns := bts[:i]
	return *(*string)(unsafe.Pointer(&ns))
}

func CopyStr(dest []byte, src string, index int, len int) int {
	copy(dest[index:index+len], src)
	index += len
	return index
}

func CopyByte(dest []byte, src byte, index int) int {
	dest[index] = src
	index++
	return index
}

func FormatTime(time time.Time) string {
	s := time.Format("060102150405")
	return s + "032+"
}

// ToTPUDHISlices 拆分为长短信切片
// 纯ASCII内容的拆分 pkgLen = 160
// 含中文内容的拆分   pkgLen = 140
func ToTPUDHISlices(content []byte, pkgLen int) (rt [][]byte) {
	if len(content) < pkgLen {
		return [][]byte{content}
	}

	headLen := 6
	bodyLen := pkgLen - headLen
	parts := len(content) / bodyLen
	tailLen := len(content) % bodyLen
	if tailLen != 0 {
		parts++
	}
	// 分片消息组的标识，用于收集组装消息
	groupId := byte(time.Now().UnixNano() & 0xff)
	var part []byte
	for i := 0; i < parts; i++ {
		if i != parts-1 {
			part = make([]byte, pkgLen)
		} else {
			// 最后一片
			part = make([]byte, 6+tailLen)
		}
		part[0], part[1], part[2] = 0x05, 0x00, 0x03
		part[3] = groupId
		part[4], part[5] = byte(parts), byte(i+1)
		if i != parts-1 {
			copy(part[6:pkgLen], content[bodyLen*i:bodyLen*(i+1)])
		} else {
			copy(part[6:], content[0:tailLen])
		}
		rt = append(rt, part)
	}
	return rt
}
