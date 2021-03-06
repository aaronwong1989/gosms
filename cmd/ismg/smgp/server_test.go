package main

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

	"github.com/aaronwong1989/gosms/codec/smgp"
	"github.com/aaronwong1989/gosms/comm"
	"github.com/aaronwong1989/gosms/comm/yml_config"
)

func init() {
	rand.Seed(time.Now().Unix()) // 随机种子
	smgp.Conf = yml_config.CreateYamlFactory("smgp.yaml")
	dc := smgp.Conf.GetInt("data-center-id")
	wk := smgp.Conf.GetInt("worker-id")
	smgwId := smgp.Conf.GetString("smgw-id")
	smgp.Seq32 = comm.NewCycleSequence(int32(dc), int32(wk))
	smgp.Seq80 = comm.NewBcdSequence(smgwId)
}

var (
	pool      = goroutine.Default()
	counterMt int64
	counterRt int64
	counterAt int64
	wg        sync.WaitGroup
	mtChan    = make(chan struct{}, 1)
	dlyChan   = make(chan struct{}, 1)
	readChan  = make(chan struct{}, 1)
	termChan  = make(chan struct{})

	clients  = 1
	duration = time.Second * 10
	// addr = "10.211.55.13:9000"
	addr = ":9100"
)

func TestClient(t *testing.T) {
	wg.Add(1)
	defer func() {
		pool.Release()
		logResult(t)
	}()

	for i := 0; i < clients; i++ {
		runClient(t)
	}
	time.Sleep(duration)

	// 停掉发送
	dlyChan <- struct{}{}
	mtChan <- struct{}{}
	// 停掉接收
	time.Sleep(100 * time.Millisecond)
	readChan <- struct{}{}
	time.Sleep(100 * time.Millisecond)
	// 发送断开连接报文
	wg.Done()
	<-termChan
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

func runClient(t *testing.T) {
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
			panic("登录失败，程序退出!")
		}

		_ = pool.Submit(func() {
			for s := true; s; {
				select {
				case <-mtChan:
					s = false
					t.Logf("接收到 mtChan 的停止信号")
				default:
					s = sendMt(t, c)
				}
			}
		})

		_ = pool.Submit(func() {
			for s := true; s; {
				select {
				case <-dlyChan:
					s = false
					t.Logf("接收到 dlyChan 的停止信号")
				default:
					s = sendDelivery(t, c)
					time.Sleep(time.Millisecond * 50)
				}
			}
		})

		_ = pool.Submit(func() {
			for s := true; s; {
				select {
				case <-readChan:
					s = false
					t.Logf("接收到 readChan 的停止信号")
				default:
					s = readResp(t, c)
				}
			}
		})

		wg.Wait()
		terminate(t, c)
		termChan <- struct{}{}
	}(t)
}

func login(t *testing.T, c net.Conn) bool {
	con := smgp.NewLogin()
	t.Logf(">>>: %s", con)
	i, _ := c.Write(con.Encode())
	assert.True(t, uint32(i) == con.PacketLength)

	pl := smgp.LoginRespLen
	resp := make([]byte, pl)
	i, _ = c.Read(resp)
	assert.True(t, i == pl)

	header := &smgp.MessageHeader{}
	err := header.Decode(resp)
	if err != nil {
		return false
	}
	rep := &smgp.LoginResp{}
	err = rep.Decode(header, resp[smgp.HeadLength:])
	if err != nil {
		return false
	}
	t.Logf("<<<: %s", rep)
	return 0 == rep.Status()
}

func sendMt(t *testing.T, c net.Conn) bool {
	mts := smgp.NewSubmit([]string{"13100001111"}, fmt.Sprintf("hello world! %d", rand.Uint64()), smgp.MtOptions{})
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
	header := &smgp.MessageHeader{}
	_ = header.Decode(bytes)
	l := int(header.PacketLength - 12)
	bytes = make([]byte, l)
	l, err = c.Read(bytes)
	if err != nil {
		t.Errorf("%v", err)
		return false
	}
	if header.RequestId == smgp.CmdSubmitResp {
		csr := &smgp.SubmitResp{}
		err := csr.Decode(header, bytes)
		if err != nil {
			t.Errorf("%v", err)
			return false
		} else {
			atomic.AddInt64(&counterMt, 1)
			t.Logf("<<< %s", csr)
		}
	} else if header.RequestId == smgp.CmdDeliver {
		dly := &smgp.Deliver{}
		err := dly.Decode(header, bytes)
		if err != nil {
			t.Errorf("%v", err)
			return false
		} else {
			// 状态报告计数
			if dly.IsReport() {
				atomic.AddInt64(&counterRt, 1)
			}
			t.Logf("<<< %s", dly)
		}
	} else if header.RequestId == smgp.CmdActiveTestResp {
		at := (*smgp.ActiveTest)(header)
		t.Logf("<<< %s", header)
		ats := at.ToResponse(0).(*smgp.ActiveTestResp)
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
	dly := smgp.NewDeliver("13700001111", "123", "hello word 中国")
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
	term := smgp.NewExit()
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
	h := &smgp.MessageHeader{}
	err = h.Decode(bytes)
	if err != nil {
		t.Errorf("%v", err)
	}
	t.Logf("<<< %s", term)
}
