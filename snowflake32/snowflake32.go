package snowflake32

import (
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"
)

// Snowflake 24小时内不会重复的雪花序号生成器
// 构成为: 0 | seconds 17 bit | datacenter 2 bit | worker 3 bit| sequence 9 bit
// 最大支持32个节点，单节点TPS不超过512，超过则会阻塞程序到下一秒再返回序号，仅能用于特殊场景
// seconds占用17bits是因为一天86400秒占用17bits
type Snowflake struct {
	sync.Mutex       // 锁
	seconds    int32 // 时间戳 ，截止到午夜0点的秒数
	datacenter int32 // 数据中心机房id, 取值范围范围：0-4
	worker     int32 // 工作节点, 取值范围范围：0-8
	sequence   int32 // 序列号
}

const (
	sequenceMask    = int32(0x01ff)                              // 最大值为9个1
	datacenterBits  = uint(2)                                    // 数据中心id所占位数
	workerBits      = uint(3)                                    // 机器id所占位数
	sequenceBits    = uint(9)                                    // 序列所占的位数
	workerShift     = sequenceBits                               // 机器id左移位数
	datacenterShift = sequenceBits + workerBits                  // 数据中心id左移位数
	timestampShift  = sequenceBits + workerBits + datacenterBits // 时间戳左移位数
)

// NewSnowflake d for datacenter-id, w for worker-id
func NewSnowflake(d int32, w int32) *Snowflake {
	return &Snowflake{datacenter: d, worker: w}
}

func (s *Snowflake) NextVal() int32 {
	s.Lock()
	defer s.Unlock()
	now := passedSeconds() // 获得当前秒
	if s.seconds == now {
		// 当同一时间戳（精度：秒）下次生成id会增加序列号
		s.sequence = (s.sequence + 1) & sequenceMask
		if s.sequence == 0 {
			// 如果当前序列超出9bit长度，则需要等待下一秒
			// 下一秒将使用sequence:0
			for now <= s.seconds {
				time.Sleep(time.Microsecond)
				now = passedSeconds()
			}
		}
	} else {
		// 不同时间戳（精度：秒）下直接使用序列号：0
		s.sequence = 0
	}
	s.seconds = now
	r := (s.seconds << timestampShift) | (s.datacenter << datacenterShift) | (s.worker << workerShift) | (s.sequence)
	return r
}

func (s *Snowflake) String() string {
	return fmt.Sprintf("%d:%d:%d:%d", s.seconds, s.datacenter, s.worker, s.sequence)
}

func passedSeconds() int32 {
	t := time.Now()
	return int32(t.Hour()*3600 + t.Minute()*60 + t.Second())
}

// Telecomflake 24小时内不会重复的雪花序号生成器
// BCD 4bit编码，用4bit表示0-9的数字
type Telecomflake struct {
	sync.Mutex        // 锁
	worker     []byte // SMGW代码：3 字节（ BCD 码），6位十进制数字的字符串
	timestamp  string // 时间：4 字节（ BCD 码），格式为 MMDDHHMM（月日时分）
	sequence   int32  // 序列号：3 字节（ BCD 码），取值范围为 000000 999999 ，从 0 开始，顺序累加，步长为 1 循环使用。
}

const telSeqMax = 1000000

func NewTelecomflake(w string) *Telecomflake {
	// check
	for _, s := range w {
		if byte(s) > '9' || byte(s) < '0' {
			w = "000000"
			break
		}
	}
	w = "000000" + w
	w = w[len(w)-6:]

	ret := &Telecomflake{}
	ret.worker = StoBcd(w)
	return ret
}

func (tf *Telecomflake) NextSeq() []byte {
	tf.Lock()
	defer tf.Unlock()
	mi := time.Now().Format("01021504")
	if tf.timestamp == mi {
		// 超过每分钟telSeqMax后序号会重复
		tf.sequence = (tf.sequence + 1) % telSeqMax
	} else {
		tf.sequence = 0
	}
	tf.timestamp = mi
	seq := make([]byte, 10)
	copy(seq[0:3], tf.worker)
	copy(seq[3:7], StoBcd(tf.timestamp))
	copy(seq[7:10], StoBcd(IntToFixStr(int64(tf.sequence), 6)))
	return seq
}

func IntToFixStr(i int64, l int) string {
	si := strconv.FormatInt(i, 10)
	if len(si) == l {
		return si
	} else {
		var sb strings.Builder
		sb.Grow(l)
		for i := 0; i < l; i++ {
			sb.WriteByte('0')
		}
		sb.WriteString(si)
		si = sb.String()
		return si[len(si)-l:]
	}
}

func StoBcd(w string) []byte {
	var wb []byte
	var h, l byte
	for i, c := range []byte(w) {
		// index 为偶数的作为高4bit
		if i&0x1 == 0 {
			h = c - '0'
			if h > 9 {
				h = 9
			}
		} else {
			// index 为奇数的作为低4bit
			l = c - '0'
			if l > 9 {
				l = 9
			}
		}
		// 每两个字符构成一个字节
		if i&0x1 == 1 {
			wb = append(wb, h<<4|l)
			h, l = 0, 0
		}
	}
	if h != 0 {
		wb = append(wb, h<<4)
	}
	return wb
}

func BcdToString(bcd []byte) string {
	var sb strings.Builder
	sb.Grow(2 * len(bcd))
	for _, b := range bcd {
		c := b >> 4
		if c > 9 {
			c = 9
		}
		sb.WriteByte(c + '0')
		c = b & 0x0f
		if c > 9 {
			c = 9
		}
		sb.WriteByte(c + '0')
	}
	return sb.String()
}
