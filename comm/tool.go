package comm

import (
	"bufio"
	"fmt"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"time"
	"unsafe"

	"github.com/panjf2000/gnet/v2"
	"golang.org/x/text/encoding/unicode"
	"golang.org/x/text/transform"

	"github.com/aaronwong1989/gosms/comm/logging"
)

var log = logging.GetDefaultLogger()

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

// TakeBytes 消费一定字节数的数据
func TakeBytes(c gnet.Conn, bytes int) []byte {
	if c.InboundBuffered() < bytes {
		return nil
	}
	frame, err := c.Peek(bytes)
	if err != nil {
		log.Errorf("[%-9s] decode error: %v", "OnTraffic", err)
		return nil
	}
	_, err = c.Discard(bytes)
	if err != nil {
		log.Errorf("[%-9s] decode error: %v", "OnTraffic", err)
		return nil
	}
	return frame
}

// Ucs2Encode Encode to UCS2.
func Ucs2Encode(s string) []byte {
	e := unicode.UTF16(unicode.BigEndian, unicode.IgnoreBOM)
	ucs, _, err := transform.Bytes(e.NewEncoder(), []byte(s))
	if err != nil {
		return nil
	}
	return ucs
}

// Ucs2Decode Decode from UCS2.
func Ucs2Decode(ucs2 []byte) string {
	e := unicode.UTF16(unicode.BigEndian, unicode.IgnoreBOM)
	bts, _, err := transform.Bytes(e.NewDecoder(), ucs2)
	if err != nil {
		return ""
	}
	return TrimStr(bts)
}

func LogHex(level logging.Level, model string, bts []byte) {
	msg := fmt.Sprintf("[OnTraffic] Hex %s: %x", model, bts)
	if level == logging.DebugLevel {
		log.Debugf(msg)
	} else if level == logging.ErrorLevel {
		log.Errorf(msg)
	} else if level == logging.WarnLevel {
		log.Warnf(msg)
	} else {
		log.Infof(msg)
	}
}

func RandNum(min, max int32) int {
	return rand.Intn(int(max-min)) + int(min)
}

// DiceCheck 投概率骰子，得到结果比给定数字大则返回true，否则返回false
func DiceCheck(prob float64) bool {
	return float64(rand.Intn(10000))/10000.0 > prob
}

// SavePid 在程序执行的当前目录生成pid文件
func SavePid(f string) string {
	file, err := os.OpenFile(f, os.O_WRONLY|os.O_CREATE, 0600)
	if err != nil {
		log.Errorf("%v", err)
	}
	pid := fmt.Sprintf("%d", os.Getpid())

	writer := bufio.NewWriter(file)
	_, _ = writer.WriteString(pid)
	defer func(file *os.File, writer *bufio.Writer) {
		_ = writer.Flush()
		_ = file.Close()
	}(file, writer)

	return pid
}

// StartMonitor 开启pprof，监听请求
func StartMonitor(port int) {
	go func() {
		addr := strconv.Itoa(port + 1)
		log.Infof("[Pprof    ] http://localhost:%s/debug/pprof/", addr)
		if err := http.ListenAndServe(":"+addr, nil); err != nil {
			log.Infof("start pprof failed on %s", addr)
		}
	}()
}
