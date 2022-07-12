package cmcc

import (
	"flag"
	"fmt"
	"net/http"
	_ "net/http/pprof"
	"strconv"
	"sync"
	"time"

	"github.com/panjf2000/ants/v2"
	"github.com/panjf2000/gnet/v2"
	"github.com/panjf2000/gnet/v2/pkg/pool/goroutine"

	"sms-vgateway/comm"
	"sms-vgateway/comm/logging"
)

var log = logging.GetDefaultLogger()

type Server struct {
	gnet.BuiltinEventEngine
	engine    gnet.Engine
	protocol  string
	address   string
	multicore bool
	pool      *goroutine.Pool
	conMap    sync.Map
	window    chan struct{}
}

func StartServer() {
	var port int
	var multicore bool
	flag.IntVar(&port, "port", 9000, "--port 9000")
	flag.BoolVar(&multicore, "multicore", true, "--multicore=true")
	flag.Parse()

	log.Infof("current pid is %s.", comm.SavePid("cmcc.pid"))
	// 定义异步工作Go程池
	options := ants.Options{
		ExpiryDuration:   time.Minute,      // 1 分钟内不被使用的worker会被清除
		Nonblocking:      false,            // 如果为true,worker池满了后提交任务会直接返回nil
		MaxBlockingTasks: Conf.MaxPoolSize, // blocking模式有效，否则worker池满了后提交任务会直接返回nil
		PreAlloc:         false,
		PanicHandler: func(e interface{}) {
			log.Errorf("%v", e)
		},
	}
	pool, _ := ants.NewPool(Conf.MaxPoolSize, ants.WithOptions(options))
	defer pool.Release()

	ss := &Server{
		protocol:  "tcp",
		address:   fmt.Sprintf(":%d", port),
		multicore: multicore,
		pool:      pool,
		window:    make(chan struct{}, Conf.ReceiveWindowSize), // 用通道控制消息接收窗口
	}

	startMonitor(port)

	err := gnet.Run(ss, ss.protocol+"://"+ss.address, gnet.WithMulticore(multicore), gnet.WithTicker(true))
	log.Errorf("server(%s://%s) exits with error: %v", ss.protocol, ss.address, err)
}

// 开启pprof，监听请求
func startMonitor(port int) {
	go func() {
		addr := strconv.Itoa(port + 1)
		log.Infof("[Pprof    ] http://localhost:%s/debug/pprof/", addr)
		if err := http.ListenAndServe(":"+addr, nil); err != nil {
			log.Infof("start pprof failed on %s", addr)
		}
	}()
}

func (s *Server) OnBoot(eng gnet.Engine) (action gnet.Action) {
	log.Infof("[%-9s] running server on %s with multi-core=%t", "OnBoot", fmt.Sprintf("%s://%s", s.protocol, s.address), s.multicore)
	s.engine = eng
	return
}

func (s *Server) OnShutdown(eng gnet.Engine) {
	log.Warnf("[%-9s] shutdown server %s ...", "OnShutdown", fmt.Sprintf("%s://%s", s.protocol, s.address))
	for eng.CountConnections() > 0 {
		log.Warnf("[%-9s] active connections is %d, waiting...", eng.CountConnections())
		time.Sleep(10 * time.Millisecond)
	}
	log.Warnf("[%-9s] shutdown server %s completed!", "OnShutdown", fmt.Sprintf("%s://%s", s.protocol, s.address))
}

func (s *Server) OnOpen(c gnet.Conn) (out []byte, action gnet.Action) {
	if s.countConn() >= Conf.MaxCons {
		log.Warnf("[%-9s] [%v<->%v] FLOW CONTROL：connections threshold reached, closing new connection...", "OnOpen", c.RemoteAddr(), c.LocalAddr())
		return nil, gnet.Close
	} else if len(s.window) == Conf.ReceiveWindowSize {
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
	s.conMap.Delete(c.RemoteAddr().String())
	return
}

func (s *Server) OnTraffic(c gnet.Conn) (action gnet.Action) {
	header := getHeader(c)
	if header == nil || header.TotalLength < 12 || header.TotalLength > 512 {
		log.Warnf("[%-9s] [%v<->%v] decode error, header: %s, close session...", "OnTraffic", c.RemoteAddr(), c.LocalAddr(), header)
		return gnet.Close
	}
	action = checkReceiveWindow(s, c, header)
	if action == gnet.Close {
		return action
	}

	switch header.CommandId {
	case CMPP_CONNECT:
		return handleConnect(s, c, header)
	case CMPP_CONNECT_RESP:
		return gnet.None
	case CMPP_SUBMIT:
		return handleSubmit(s, c, header)
	case CMPP_SUBMIT_RESP:
		return gnet.None
	case CMPP_DELIVER:
		return handleDelivery(s, c, header)
	case CMPP_DELIVER_RESP:
		return gnet.None
	case CMPP_ACTIVE_TEST:
		return handActive(s, c, header)
	case CMPP_ACTIVE_TEST_RESP:
		return handActiveResp(c, header)
	case CMPP_TERMINATE:
		return handleTerminate(s, c, header)
	case CMPP_TERMINATE_RESP:
		return handleTerminateResp(s, c, header)
	}

	return gnet.None
}

func (s *Server) OnTick() (delay time.Duration, action gnet.Action) {
	log.Infof("[%-9s] %d active connections.", "OnTick", s.activeCons())
	s.conMap.Range(func(key, value interface{}) bool {
		addr := key.(string)
		con, ok := value.(gnet.Conn)
		if ok {
			_ = s.pool.Submit(func() {
				at := NewActiveTest()
				err := con.AsyncWrite(at.Encode(), nil)
				if err == nil {
					log.Infof("[%-9s] >>> %s to %s", "OnTick", at, addr)
				} else {
					log.Errorf("[%-9s] >>> CMPP_ACTIVE_TEST to %s, error: %v", "OnTick", addr, err)
				}
			})
		}
		return true
	})
	return Conf.ActiveTestDuration, gnet.None
}

func (s *Server) countConn() int {
	counter := 0
	s.conMap.Range(func(key, value interface{}) bool {
		counter++
		return true
	})
	return counter
}

func (s *Server) activeCons() int {
	return s.engine.CountConnections()
}

func handleConnect(s *Server, c gnet.Conn, header *MessageHeader) gnet.Action {
	frame := comm.TakeBytes(c, 39-HEAD_LENGTH)
	comm.LogHex(logging.DebugLevel, "Connect", frame)

	connect := &Connect{}
	err := connect.Decode(header, frame)
	if err != nil {
		log.Errorf("[%-9s] CMPP_CONNECT ERROR: %v", "OnTraffic", err)
		return gnet.Close
	}

	log.Infof("[%-9s] <<< %s", "OnTraffic", connect)
	resp := connect.ToResponse(0).(*ConnectResp)
	if resp.Status() != 0 {
		log.Errorf("[%-9s] CMPP_CONNECT ERROR: Auth Error, status=(%d,%s)", "OnTraffic", resp.Status(), ConnectStatusMap[resp.Status()])
	}

	// send cmpp_connect_resp async
	_ = s.pool.Submit(func() {
		err = c.AsyncWrite(resp.Encode(), func(c gnet.Conn) error {
			log.Infof("[%-9s] >>> %s", "OnTraffic", resp)
			if resp.Status() == 0 {
				s.conMap.Store(c.RemoteAddr().String(), c)
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

func handleTerminate(s *Server, c gnet.Conn, header *MessageHeader) gnet.Action {
	log.Infof("[%-9s] <<< %s", "OnTraffic", header)
	resp := NewTerminateResp(header.SequenceId)
	// send cmpp_connect_resp async
	_ = s.pool.Submit(func() {
		err := c.AsyncWrite(resp.Encode(), func(c gnet.Conn) error {
			log.Infof("[%-9s] >>> %s", "OnTraffic", resp)
			s.conMap.Delete(c.RemoteAddr().String())
			_ = c.Close()
			return nil
		})
		if err != nil {
			log.Errorf("[%-9s] CMPP_TERMINATE_RESP ERROR: %v", "OnTraffic", err)
		}
	})
	return gnet.None
}

func handleTerminateResp(s *Server, c gnet.Conn, header *MessageHeader) gnet.Action {
	log.Infof("[%-9s] <<< %s", "OnTraffic", header)
	log.Infof("[%-9s] closing connection [%v<-->%v]", "OnTraffic", c.RemoteAddr(), c.LocalAddr())
	s.conMap.Delete(c.RemoteAddr().String())
	_ = c.Flush()
	_ = c.Close()
	return gnet.Close
}

// 处理上行消息
func handleDelivery(s *Server, c gnet.Conn, header *MessageHeader) gnet.Action {
	// check connect
	_, ok := s.conMap.Load(c.RemoteAddr().String())
	if !ok {
		log.Warnf("[%-9s] unLogin connection: %s, closing...", "OnTraffic", c.RemoteAddr())
		return gnet.Close
	}

	frame := comm.TakeBytes(c, int(header.TotalLength-HEAD_LENGTH))
	comm.LogHex(logging.DebugLevel, "Delivery", frame)
	dly := &Delivery{}
	err := dly.Decode(header, frame)
	if err != nil {
		log.Errorf("[%-9s] CMPP_DELIVERY ERROR: %v", "OnTraffic", err)
		return gnet.Close
	}
	log.Debugf("[%-9s] <<< %s", "OnTraffic", dly)
	// handle message async
	_ = s.pool.Submit(func() {
		// 模拟消息处理耗时
		processTime := time.Duration(comm.RandNum(Conf.MinSubmitRespMs, Conf.MaxSubmitRespMs))
		time.Sleep(processTime * time.Millisecond)

		rtCode := uint32(0)
		if comm.DiceCheck(Conf.SuccessRate) {
			// 失败消息的返回码
			rtCode = 9
		}
		resp := dly.ToResponse(rtCode).(*DeliveryResp)
		// 发送响应
		err := c.AsyncWrite(resp.Encode(), func(c gnet.Conn) error {
			log.Debugf("[%-9s] >>> %s", "OnTraffic", resp)
			return nil
		})
		if err != nil {
			log.Errorf("[%-9s] CMPP_DELIVERY_RESP ERROR: %v", "OnTraffic", err)
		}
	})
	return gnet.None
}

func handleSubmit(s *Server, c gnet.Conn, header *MessageHeader) gnet.Action {
	// check connect
	_, ok := s.conMap.Load(c.RemoteAddr().String())
	if !ok {
		log.Warnf("[%-9s] unLogin connection: %s, closing...", "OnTraffic", c.RemoteAddr())
		return gnet.Close
	}

	frame := comm.TakeBytes(c, int(header.TotalLength-HEAD_LENGTH))
	comm.LogHex(logging.DebugLevel, "Submit", frame)
	sub := &Submit{}
	err := sub.Decode(header, frame)
	if err != nil {
		log.Errorf("[%-9s] CMPP_SUBMIT ERROR: %v", "OnTraffic", err)
		return gnet.Close
	}
	log.Debugf("[%-9s] <<< %s", "OnTraffic", sub)
	// handle message async
	_ = s.pool.Submit(mtAsyncHandler(s, c, sub))
	return gnet.None
}

func mtAsyncHandler(s *Server, c gnet.Conn, sub *Submit) func() {
	return func() {
		// 采用通道控制消息收发速度,向通道发送信号
		s.window <- struct{}{}
		defer func() {
			// defer函数消费信号，确保每个消息的信号最终都会被消费
			<-s.window
		}()

		// 模拟消息处理耗时，可配置
		processTime := time.Duration(0)
		if Conf.MinSubmitRespMs > 0 && Conf.MaxSubmitRespMs > Conf.MinSubmitRespMs {
			processTime = time.Duration(comm.RandNum(Conf.MinSubmitRespMs, Conf.MaxSubmitRespMs))
			time.Sleep(processTime * time.Millisecond)
		}

		rtCode := uint32(0)
		if comm.DiceCheck(Conf.SuccessRate) {
			// 失败消息的返回码
			rtCode = 13
		}
		resp := sub.ToResponse(rtCode).(*SubmitResp)
		// 发送响应
		err := c.AsyncWrite(resp.Encode(), func(c gnet.Conn) error {
			log.Debugf("[%-9s] >>> %s", "OnTraffic", resp)
			return nil
		})
		if err != nil {
			log.Errorf("[%-9s] CMPP_SUBMIT_RESP ERROR: %v", "OnTraffic", err)
		}

		// 发送状态报告
		if resp.Result() == 0 {
			_ = s.pool.Submit(reportAsyncSender(c, sub, resp.MsgId(), processTime))
		}
	}
}

func reportAsyncSender(c gnet.Conn, sub *Submit, msgId uint64, wait time.Duration) func() {
	return func() {
		if comm.DiceCheck(100) {
			return
		}
		dly := sub.ToDeliveryReport(msgId)
		// 模拟状态报告发送前的耗时
		if Conf.FixReportRespMs > 0 {
			processTime := wait + time.Duration(Conf.FixReportRespMs)
			time.Sleep(processTime * time.Millisecond)
		}
		// 发送状态报告
		err := c.AsyncWrite(dly.Encode(), func(c gnet.Conn) error {
			log.Debugf("[%-9s] >>> %s", "OnTraffic", dly)
			return nil
		})
		if err != nil {
			log.Errorf("[%-9s] CMPP_DELIVERY_REPORT ERROR: %v", "OnTraffic", err)
		}
	}
}

func handActive(s *Server, c gnet.Conn, header *MessageHeader) (action gnet.Action) {
	respHeader := &MessageHeader{TotalLength: 13, CommandId: CMPP_ACTIVE_TEST_RESP, SequenceId: header.SequenceId}
	resp := &ActiveTestResp{MessageHeader: respHeader}
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

func handActiveResp(c gnet.Conn, header *MessageHeader) (action gnet.Action) {
	if c.InboundBuffered() >= 1 {
		_, _ = c.Discard(1)
	}
	log.Infof("[%-9s] <<< %s from %s", "OnTraffic", &ActiveTestResp{MessageHeader: header}, c.RemoteAddr())
	return gnet.None
}

func getHeader(c gnet.Conn) *MessageHeader {
	frame := comm.TakeBytes(c, HEAD_LENGTH)
	if frame == nil {
		return nil
	}
	comm.LogHex(logging.DebugLevel, "Header", frame)

	header := MessageHeader{}
	err := header.Decode(frame)
	if err != nil {
		log.Errorf("[%-9s] decode error: %v", "OnTraffic", err)
		return nil
	}
	return &header
}

func checkReceiveWindow(s *Server, c gnet.Conn, header *MessageHeader) gnet.Action {
	if len(s.window) == Conf.ReceiveWindowSize && header.CommandId == CMPP_SUBMIT {
		log.Warnf("[%-9s] FLOW CONTROL：receive window threshold reached.", "OnTraffic")
		l := int(header.TotalLength - HEAD_LENGTH)
		discard, err := c.Discard(l)
		if err != nil || discard != l {
			return gnet.Close
		}
		sub := &Submit{}
		sub.MessageHeader = header
		resp := sub.ToResponse(8).(*SubmitResp)
		// 发送响应
		err = c.AsyncWrite(resp.Encode(), func(c gnet.Conn) error {
			log.Debugf("[%-9s] >>> %s", "OnTraffic", resp)
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
