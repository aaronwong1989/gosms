package main

// export CMCC_CONF_PATH="/Users/huangzhonghui/.cmcc.yaml"
// export GNET_LOGGING_LEVEL=-1
// export GNET_LOGGING_FILE="/Users/huangzhonghui/logs/sms.log"

import (
	"flag"
	"fmt"
	"math/rand"
	"sync"
	"time"

	"github.com/panjf2000/gnet/v2"
	"github.com/panjf2000/gnet/v2/pkg/logging"
	"github.com/panjf2000/gnet/v2/pkg/pool/goroutine"

	cmcc "sms-vgateway/cmcc/protocol"
)

var log = logging.GetDefaultLogger()

func main() {
	var port int
	var multicore bool

	// Example command: go run server.go --port 9000 --multicore=true
	flag.IntVar(&port, "port", 9000, "--port 9000")
	flag.BoolVar(&multicore, "multicore", true, "--multicore=true")
	flag.Parse()
	pool := goroutine.Default()
	defer pool.Release()

	ss := &Server{
		protocol:         "tcp",
		address:          fmt.Sprintf(":%d", port),
		multicore:        multicore,
		pool:             pool,
		connectedSockets: make(map[string]gnet.Conn),
		window:           make(chan struct{}, cmcc.Conf.ReceiveWindowSize), // 用通道控制消息接收窗口
	}
	log.Warnf("server(%s://%s) stating....", ss.protocol, ss.address)
	err := gnet.Run(ss, ss.protocol+"://"+ss.address, gnet.WithMulticore(multicore), gnet.WithTicker(true))
	log.Errorf("server(%s://%s) exits with error: %v", ss.protocol, ss.address, err)
}

type Server struct {
	sync.RWMutex
	gnet.BuiltinEventEngine
	engine           gnet.Engine
	protocol         string
	address          string
	multicore        bool
	pool             *goroutine.Pool
	connectedSockets map[string]gnet.Conn
	window           chan struct{}
}

func (s *Server) OnBoot(eng gnet.Engine) (action gnet.Action) {
	log.Infof("[%-9s] running server on %s with multi-core=%t", "OnBoot", fmt.Sprintf("%s://%s", s.protocol, s.address), s.multicore)
	s.engine = eng
	return
}

func (s *Server) OnShutdown(eng gnet.Engine) {
	log.Infof("[%-9s] shutdown server %s ...", "OnShutdown", fmt.Sprintf("%s://%s", s.protocol, s.address))
	for eng.CountConnections() > 0 {
		log.Infof("[%-9s] active connections is %d, waiting...", eng.CountConnections())
		time.Sleep(10 * time.Millisecond)
	}
	log.Infof("[%-9s] shutdown server %s completed!", "OnShutdown", fmt.Sprintf("%s://%s", s.protocol, s.address))
}

func (s *Server) OnOpen(c gnet.Conn) (out []byte, action gnet.Action) {
	if s.countConn() >= cmcc.Conf.MaxCons {
		log.Warnf("[%-9s] [%v<->%v] FLOW CONTROL：connections threshold reached, closing new connection...", "OnOpen", c.RemoteAddr(), c.LocalAddr())
		return nil, gnet.Close
	} else if len(s.window) == cmcc.Conf.ReceiveWindowSize {
		log.Warnf("[%-9s] [%v<->%v] FLOW CONTROL：receive window threshold reached, closing new connection...", "OnOpen", c.RemoteAddr(), c.LocalAddr())
		// 已达到窗口时，拒绝新的连接
		return nil, gnet.Close
	} else {
		log.Infof("[%-9s] [%v<->%v] activeCons=%d.", "OnOpen", c.RemoteAddr(), c.LocalAddr(), s.activeCons())
		return
	}
}

func (s *Server) OnClose(c gnet.Conn, e error) (action gnet.Action) {
	log.Warnf("[%-9s] [%v<->%v] activeCons=%d, reason=%v.", "OnClose", c.RemoteAddr(), c.LocalAddr(), s.activeCons(), e)
	s.Lock()
	defer s.Unlock()
	delete(s.connectedSockets, c.RemoteAddr().String())
	return
}

func (s *Server) OnTraffic(c gnet.Conn) (action gnet.Action) {
	header := getHeader(c)
	if header == nil {
		log.Warnf("[%-9s] [%v<->%v] decode error, close session...", "OnTraffic", c.RemoteAddr(), c.LocalAddr())
		return gnet.Close
	}
	action = checkReceiveWindow(s, c, header)
	if action == gnet.Close {
		return action
	}

	switch header.CommandId {
	case cmcc.CMPP_CONNECT:
		return handleConnect(s, c, header)
	case cmcc.CMPP_CONNECT_RESP:
	case cmcc.CMPP_SUBMIT:
		return handleSubmit(s, c, header)
	case cmcc.CMPP_SUBMIT_RESP:
	case cmcc.CMPP_DELIVER:
	case cmcc.CMPP_DELIVER_RESP:
	case cmcc.CMPP_ACTIVE_TEST:
		return handActive(s, c, header)
	case cmcc.CMPP_ACTIVE_TEST_RESP:
		return handActiveResp(c, header)
	case cmcc.CMPP_TERMINATE:
	case cmcc.CMPP_TERMINATE_RESP:
	}

	return gnet.None
}

func (s *Server) OnTick() (delay time.Duration, action gnet.Action) {
	log.Infof("[%-9s] %d active connections.", "OnTick", s.activeCons())
	s.Lock()
	defer s.Unlock()
	for addr, con := range s.connectedSockets {
		_ = s.pool.Submit(func() {
			at := cmcc.NewActiveTest()
			err := con.AsyncWrite(at.Encode(), nil)
			if err == nil {
				log.Infof("[%-9s] >>> %s to %s", "OnTick", at, addr)
			} else {
				log.Errorf("[%-9s] >>> CMPP_ACTIVE_TEST to %s, error: %v", "OnTick", addr, err)
			}
		})
	}
	return cmcc.Conf.ActiveTestDuration, gnet.None
}

func (s *Server) countConn() int {
	s.RLock()
	defer s.RUnlock()
	return len(s.connectedSockets)
}

func handleConnect(s *Server, c gnet.Conn, header *cmcc.MessageHeader) gnet.Action {
	frame := take(c, 39-cmcc.HEAD_LENGTH)
	logHex(logging.DebugLevel, "Connect", frame)

	connect := &cmcc.CmppConnect{}
	err := connect.Decode(header, frame)
	if err != nil {
		log.Errorf("[%-9s] CMPP_CONNECT ERROR: %v", "OnTraffic", err)
		return gnet.Close
	}

	log.Infof("[%-9s] <<< %s", "OnTraffic", connect)
	resp := connect.ToResponse()
	if resp.Status != 0 {
		log.Errorf("[%-9s] CMPP_CONNECT ERROR: Auth Error, Status=(%d,%s)", "OnTraffic", resp.Status, cmcc.ConnectStatusMap[resp.Status])
	}

	// send cmpp_connect_resp async
	_ = s.pool.Submit(func() {
		err = c.AsyncWrite(resp.Encode(), func(c gnet.Conn) error {
			log.Infof("[%-9s] >>> %s", "OnTraffic", resp)
			if resp.Status == 0 {
				s.Lock()
				defer s.Unlock()
				s.connectedSockets[c.RemoteAddr().String()] = c
			} else {
				// 客户端登录失败，关闭连接
				_ = c.Close()
			}
			return nil
		})
		if err != nil {
			log.Errorf("[%-9s] CMPP_CONNECT_RESP ERROR: %v", "OnTraffic", err)
		}
	})
	return gnet.None
}

func handleSubmit(s *Server, c gnet.Conn, header *cmcc.MessageHeader) gnet.Action {
	// check connect
	if s.connectedSockets[c.RemoteAddr().String()] == nil {
		log.Warnf("[%-9s] unLogin connection: %s, closing...", "OnTraffic", c.RemoteAddr())
		return gnet.Close
	}

	frame := take(c, int(header.TotalLength-cmcc.HEAD_LENGTH))
	logHex(logging.DebugLevel, "Submit", frame)
	sub := &cmcc.Submit{}
	err := sub.Decode(header, frame)
	if err != nil {
		log.Errorf("[%-9s] CMPP_SUBMIT ERROR: %v", "OnTraffic", err)
		return gnet.Close
	}
	log.Infof("[%-9s] <<< %s", "OnTraffic", sub)
	// handle message async
	_ = s.pool.Submit(func() {
		// 采用通道控制消息收发速度,向通道发送信号
		s.window <- struct{}{}
		defer func() {
			// defer函数消费信号，确保每个消息的信号最终都会被消费
			<-s.window
		}()
		// TODO 上述实现方式可以控制返回消息的速率，但如果发送方不按照协议控制速度，会导致线程池中的任务越来越多

		// 模拟消息处理耗时，可配置
		processTime := time.Duration(randNum(cmcc.Conf.MinSubmitRespMs, cmcc.Conf.MaxSubmitRespMs))
		time.Sleep(processTime * time.Millisecond)

		rtCode := uint32(0)
		if sub.SequenceId%cmcc.Conf.SuccessRate == 0 {
			// 失败消息的返回码
			rtCode = 13
		}
		resp := sub.ToResponse(rtCode)

		// 发送响应
		err = c.AsyncWrite(sub.ToResponse(rtCode).Encode(), func(c gnet.Conn) error {
			log.Infof("[%-9s] >>> %s", "OnTraffic", resp)
			return nil
		})
		if err != nil {
			log.Errorf("[%-9s] CMPP_SUBMIT_RESP ERROR: %v", "OnTraffic", err)
		}
	})
	return gnet.None
}

func handActive(s *Server, c gnet.Conn, header *cmcc.MessageHeader) (action gnet.Action) {
	respHeader := &cmcc.MessageHeader{TotalLength: cmcc.HEAD_LENGTH, CommandId: cmcc.CMPP_ACTIVE_TEST_RESP, SequenceId: header.SequenceId}
	resp := &cmcc.ActiveTestResp{MessageHeader: respHeader}
	// send cmpp_active_resp async
	_ = s.pool.Submit(func() {
		err := c.AsyncWrite(resp.Encode(), func(c gnet.Conn) error {
			log.Infof("[%-9s] >>> %s", "OnTraffic", resp)
			return nil
		})
		if err != nil {
			log.Errorf("[%-9s] CMPP_ACTIVE_TEST_RESP ERROR: %v", "OnTraffic", err)
		}
	})
	return gnet.None
}

func handActiveResp(c gnet.Conn, header *cmcc.MessageHeader) (action gnet.Action) {
	if c.InboundBuffered() >= 1 {
		_, _ = c.Discard(1)
	}
	log.Infof("[%-9s] <<< %s from %s", "OnTraffic", &cmcc.ActiveTestResp{MessageHeader: header}, c.RemoteAddr())
	return gnet.None
}

func getHeader(c gnet.Conn) *cmcc.MessageHeader {
	frame := take(c, cmcc.HEAD_LENGTH)
	if frame == nil {
		return nil
	}
	logHex(logging.DebugLevel, "Header", frame)

	header := cmcc.MessageHeader{}
	err := header.Decode(frame)
	if err != nil {
		log.Errorf("[%-9s] decode error: %v", "OnTraffic", err)
		return nil
	}
	return &header
}

func checkReceiveWindow(s *Server, c gnet.Conn, header *cmcc.MessageHeader) gnet.Action {
	if len(s.window) == cmcc.Conf.ReceiveWindowSize && header.CommandId == cmcc.CMPP_SUBMIT {
		log.Infof("[%-9s] FLOW CONTROL：receive window threshold reached.", "OnTraffic")
		l := int(header.TotalLength - cmcc.HEAD_LENGTH)
		discard, err := c.Discard(l)
		if err != nil || discard != l {
			return gnet.Close
		}
		sub := &cmcc.Submit{}
		sub.MessageHeader = header
		resp := sub.ToResponse(8)
		// 发送响应
		err = c.AsyncWrite(resp.Encode(), func(c gnet.Conn) error {
			log.Infof("[%-9s] >>> %s", "OnTraffic", resp)
			return nil
		})
		if err != nil {
			log.Errorf("[%-9s] CMPP_SUBMIT_RESP ERROR: %v", "OnTraffic", err)
			return gnet.Close
		}
		header.CommandId = 0
	}
	return gnet.None
}

// 消费一定字节数的数据
func take(c gnet.Conn, bytes int) []byte {
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

func logHex(level logging.Level, model string, bts []byte) {
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

func (s *Server) activeCons() int {
	return s.engine.CountConnections()
}

func randNum(min, max int) int {
	rand.Seed(time.Now().Unix()) // 随机种子
	return rand.Intn(max-min) + min
}
