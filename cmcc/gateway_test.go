package cmcc

import (
	"bufio"
	"fmt"
	"math/rand"
	"net"
	"os"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/panjf2000/gnet/v2/pkg/pool/goroutine"
	"github.com/stretchr/testify/assert"
)

var (
	pool      = goroutine.Default()
	counterMt int64
	counterRt int64
	counterAt int64
	wg        sync.WaitGroup

	clients  = 1
	duration = time.Second * 100
	// addr = "10.211.55.13:9000"
	addr = ":9000"
)

func TestClient(t *testing.T) {
	senders := 1
	receivers := 1
	wg.Add(1)

	for i := 0; i < clients; i++ {
		runClient(t, senders, receivers)
	}
	time.Sleep(duration)
	logResult(t)

	defer func() {
		pool.Release()
		wg.Done()
	}()
}

func TestLogin(t *testing.T) {
	c, err := net.Dial("tcp", addr)
	if err != nil {
		t.Errorf("%v", err)
		return
	}
	defer func(c net.Conn) {
		err := c.Close()
		if err != nil {
			t.Errorf("%v", err)
		}
	}(c)

	login(t, c)
}

func TestTerminate(t *testing.T) {
	c, err := net.Dial("tcp", addr)
	if err != nil {
		t.Errorf("%v", err)
		return
	}
	defer func(c net.Conn) {
		err := c.Close()
		if err != nil {
			t.Errorf("%v", err)
		}
	}(c)

	if !login(t, c) {
		return
	}

	terminate(t, c)
}

func runClient(t *testing.T, senders, receivers int) {
	go func(t *testing.T) {
		c, err := net.Dial("tcp", addr)
		if err != nil {
			t.Errorf("%v", err)
			return
		}
		defer func(c net.Conn) {
			err := c.Close()
			if err != nil {
				t.Errorf("%v", err)
			}
		}(c)

		if !login(t, c) {
			return
		}
		time.Sleep(time.Millisecond * 10)

		for i := 0; i < senders; i++ {
			_ = pool.Submit(func() {
				for b := true; b; {
					b = sendMt(t, c)
				}
			})
		}

		for i := 0; i < receivers; i++ {
			_ = pool.Submit(func() {
				for b := true; b; {
					b = readResp(t, c)
				}
			})
		}

		// 上行短信发送
		_ = pool.Submit(func() {
			for b := true; b; {
				b = sendDelivery(t, c)
				time.Sleep(time.Millisecond * 100)
			}
		})

		wg.Wait()

		terminate(t, c)
	}(t)
}

func login(t *testing.T, c net.Conn) bool {
	con := NewConnect()
	t.Logf(">>>: %s", con)
	i, _ := c.Write(con.Encode())
	assert.True(t, uint32(i) == con.TotalLength)

	pl := 30
	if V3() {
		pl = 33
	}
	resp := make([]byte, pl)
	i, _ = c.Read(resp)
	assert.True(t, i == pl)

	header := &MessageHeader{}
	err := header.Decode(resp)
	if err != nil {
		return false
	}
	rep := &ConnectResp{}
	err = rep.Decode(header, resp[HEAD_LENGTH:])
	if err != nil {
		return false
	}
	t.Logf("<<<: %s", rep)
	return 0 == rep.Status()
}

func sendMt(t *testing.T, c net.Conn) bool {
	mts := NewSubmit([]string{"13100001111"}, fmt.Sprintf("hello world! %d", rand.Uint64()))
	mt := mts[0]
	_, err := c.Write(mt.Encode())
	if err != nil {
		t.Errorf("%v", err)
		return false
	}
	t.Logf(">>> %s", mt)
	return true
}

func readResp(t *testing.T, c net.Conn) bool {
	bytes := make([]byte, 12)
	_, err := c.Read(bytes)
	if err != nil {
		t.Errorf("%v", err)
		return false
	}
	header := &MessageHeader{}
	_ = header.Decode(bytes)
	l := int(header.TotalLength - 12)
	bytes = make([]byte, l)
	l, err = c.Read(bytes)
	if err != nil {
		t.Errorf("%v", err)
		return false
	}
	if header.CommandId == CMPP_SUBMIT_RESP {
		csr := &SubmitResp{}
		err := csr.Decode(header, bytes)
		if err != nil {
			t.Errorf("%v", err)
			return false
		} else {
			atomic.AddInt64(&counterMt, 1)
			t.Logf("<<< %s", csr)
		}
	} else if header.CommandId == CMPP_DELIVER {
		dly := &Delivery{}
		err := dly.Decode(header, bytes)
		if err != nil {
			t.Errorf("%v", err)
			return false
		} else {
			// 状态报告计数
			if dly.RegisteredDelivery() == 1 {
				atomic.AddInt64(&counterRt, 1)
			}
			t.Logf("<<< %s", dly)
		}
	} else if header.CommandId == CMPP_ACTIVE_TEST {
		at := ActiveTest{MessageHeader: header}
		t.Logf("<<< %s", at)
		ats := at.ToResponse(0).(*ActiveTestResp)
		_, err = c.Write(ats.Encode())
		if err != nil {
			t.Errorf("%v", err)
			return false
		} else {
			atomic.AddInt64(&counterAt, 1)
			t.Logf(">>> %s", ats)
		}
	} else {
		t.Logf("<<< %s:%x", header, bytes)
	}
	return true
}

func sendDelivery(t *testing.T, c net.Conn) bool {
	dly := NewDelivery("13700001111", "hello word 中国", "", "")
	_, err := c.Write(dly.Encode())
	if err != nil {
		t.Errorf("%v", err)
		return false
	}
	t.Logf(">>> %s", dly)
	return true
}

func logResult(t *testing.T) {
	result := fmt.Sprintf("%s CounterMt=%d, CounterDl=%d, CounterAt=%d\n", time.Now().Format("2006-01-02T15:04:05.000"), counterMt, counterRt, counterAt)
	t.Logf(result)
	file, err := os.OpenFile("./test.result.txt", os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		t.Errorf("%v", err)
	}
	writer := bufio.NewWriter(file)
	_, _ = writer.WriteString(result)
	defer func(file *os.File, writer *bufio.Writer) {
		_ = writer.Flush()
		_ = file.Close()
	}(file, writer)
}

func terminate(t *testing.T, c net.Conn) {
	term := NewTerminate()
	_, err := c.Write(term.Encode())
	if err != nil {
		t.Errorf("%v", err)
		return
	}
	t.Logf(">>> %s", term)

	bytes := make([]byte, 12)
	l, err := c.Read(bytes)
	if err != nil || l != 12 {
		t.Errorf("%v", err)
	}
	err = term.Decode(bytes)
	if err != nil {
		t.Errorf("%v", err)
	}
	t.Logf("<<< %s", term)
}