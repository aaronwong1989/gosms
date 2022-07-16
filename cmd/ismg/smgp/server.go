package main

import (
	"flag"
	"fmt"
	_ "net/http/pprof"
	"sync"
	"time"

	"github.com/panjf2000/ants/v2"
	"github.com/panjf2000/gnet/v2"
	"github.com/panjf2000/gnet/v2/pkg/pool/goroutine"

	"gosms/codec/smgp"
	"gosms/comm"
	"gosms/comm/logging"
)

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
	flag.IntVar(&port, "port", 9100, "--port 9100")
	flag.BoolVar(&multicore, "multicore", true, "--multicore=true")
	flag.Parse()

	log.Infof("current pid is %s.", comm.SavePid("smgp.pid"))
	// 定义异步工作Go程池
	options := ants.Options{
		ExpiryDuration:   time.Minute,           // 1 分钟内不被使用的worker会被清除
		Nonblocking:      false,                 // 如果为true,worker池满了后提交任务会直接返回nil
		MaxBlockingTasks: smgp.Conf.MaxPoolSize, // blocking模式有效，否则worker池满了后提交任务会直接返回nil
		PreAlloc:         false,
		PanicHandler: func(e interface{}) {
			log.Errorf("%v", e)
		},
	}
	pool, _ := ants.NewPool(smgp.Conf.MaxPoolSize, ants.WithOptions(options))
	defer pool.Release()

	ss := &Server{
		protocol:  "tcp",
		address:   fmt.Sprintf(":%d", port),
		multicore: multicore,
		pool:      pool,
		window:    make(chan struct{}, smgp.Conf.ReceiveWindowSize), // 用通道控制消息接收窗口
	}

	comm.StartMonitor(port)

	err := gnet.Run(ss, ss.protocol+"://"+ss.address, gnet.WithMulticore(multicore), gnet.WithTicker(true))
	log.Errorf("server(%s://%s) exits with error: %v", ss.protocol, ss.address, err)
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
	if s.countConn() >= smgp.Conf.MaxCons {
		log.Warnf("[%-9s] [%v<->%v] FLOW CONTROL：connections threshold reached, closing new connection...", "OnOpen", c.RemoteAddr(), c.LocalAddr())
		return nil, gnet.Close
	} else if len(s.window) == smgp.Conf.ReceiveWindowSize {
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
	if header == nil || header.PacketLength < 12 || header.PacketLength > 512 {
		log.Warnf("[%-9s] [%v<->%v] decode error, header: %s, close session...", "OnTraffic", c.RemoteAddr(), c.LocalAddr(), header)
		return gnet.Close
	}
	action = checkReceiveWindow(s, c, header)
	if action == gnet.Close {
		return action
	}

	switch header.RequestId {
	case 0: // 触发限速
		return gnet.None
	case smgp.CmdLogin:
		return handleLogin(s, c, header)
	case smgp.CmdLoginResp:
		return gnet.None
	case smgp.CmdSubmit:
		return handleSubmit(s, c, header)
	case smgp.CmdSubmitResp:
		return gnet.None
	case smgp.CmdDeliver:
		return handleDelivery(s, c, header)
	case smgp.CmdDeliverResp:
		return handleDeliveryResp(s, c, header)
	case smgp.CmdActiveTest:
		return handActive(s, c, header)
	case smgp.CmdActiveTestResp:
		return handActiveResp(c, header)
	case smgp.CmdExit:
		return handleExit(s, c, header)
	case smgp.CmdExitResp:
		return handleExitResp(s, c, header)
	default:
		// 不合法包，关闭连接
		return gnet.Close
	}
}

func (s *Server) OnTick() (delay time.Duration, action gnet.Action) {
	log.Infof("[%-9s] %d active connections.", "OnTick", s.activeCons())
	s.conMap.Range(func(key, value interface{}) bool {
		addr := key.(string)
		con, ok := value.(gnet.Conn)
		if ok {
			_ = s.pool.Submit(func() {
				at := smgp.NewActiveTest()
				err := con.AsyncWrite(at.Encode(), nil)
				if err == nil {
					log.Infof("[%-9s] >>> %s to %s", "OnTick", at, addr)
				} else {
					log.Errorf("[%-9s] >>> ACTIVE_TEST to %s, error: %v", "OnTick", addr, err)
				}
			})
		}
		return true
	})
	return smgp.Conf.ActiveTestDuration, gnet.None
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

func handleLogin(s *Server, c gnet.Conn, header *smgp.MessageHeader) gnet.Action {
	frame := comm.TakeBytes(c, smgp.LoginLen-smgp.HeadLength)
	comm.LogHex(logging.DebugLevel, "Login", frame)

	connect := &smgp.Login{}
	err := connect.Decode(header, frame)
	if err != nil {
		log.Errorf("[%-9s] LOGIN ERROR: %v", "OnTraffic", err)
		return gnet.Close
	}

	log.Infof("[%-9s] <<< %s", "OnTraffic", connect)
	resp := connect.ToResponse(0).(*smgp.LoginResp)
	if resp.Status() != 0 {
		log.Errorf("[%-9s] LOGIN ERROR: Auth Error, status=(%d,%s)", "OnTraffic", resp.Status(), smgp.ConnectStatusMap[resp.Status()])
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
			log.Errorf("[%-9s] LOGIN_RESP ERROR: %v", "OnTraffic", err)
		}
	})
	return gnet.None
}

func handleExit(s *Server, c gnet.Conn, header *smgp.MessageHeader) gnet.Action {
	log.Infof("[%-9s] <<< %s", "OnTraffic", header)
	resp := smgp.NewExitResp(header.SequenceId)
	// send cmpp_connect_resp async
	_ = s.pool.Submit(func() {
		err := c.AsyncWrite(resp.Encode(), func(c gnet.Conn) error {
			log.Infof("[%-9s] >>> %s", "OnTraffic", resp)
			s.conMap.Delete(c.RemoteAddr().String())
			_ = c.Close()
			return nil
		})
		if err != nil {
			log.Errorf("[%-9s] EXIT_RESP ERROR: %v", "OnTraffic", err)
		}
	})
	return gnet.None
}

func handleExitResp(s *Server, c gnet.Conn, header *smgp.MessageHeader) gnet.Action {
	log.Infof("[%-9s] <<< %s", "OnTraffic", header)
	log.Infof("[%-9s] closing connection [%v<-->%v]", "OnTraffic", c.RemoteAddr(), c.LocalAddr())
	s.conMap.Delete(c.RemoteAddr().String())
	_ = c.Flush()
	_ = c.Close()
	return gnet.Close
}

// 处理上行消息
func handleDelivery(s *Server, c gnet.Conn, header *smgp.MessageHeader) gnet.Action {
	// check connect
	_, ok := s.conMap.Load(c.RemoteAddr().String())
	if !ok {
		log.Warnf("[%-9s] unLogin connection: %s, closing...", "OnTraffic", c.RemoteAddr())
		return gnet.Close
	}

	frame := comm.TakeBytes(c, int(header.PacketLength-smgp.HeadLength))
	comm.LogHex(logging.DebugLevel, "Deliver", frame)
	dly := &smgp.Deliver{}
	err := dly.Decode(header, frame)
	if err != nil {
		log.Errorf("[%-9s] DELIVER ERROR: %v", "OnTraffic", err)
		return gnet.Close
	}
	log.Debugf("[%-9s] <<< %s", "OnTraffic", dly)
	// handle message async
	_ = s.pool.Submit(func() {
		// 模拟消息处理耗时
		processTime := time.Duration(comm.RandNum(smgp.Conf.MinSubmitRespMs, smgp.Conf.MaxSubmitRespMs))
		time.Sleep(processTime * time.Millisecond)

		rtCode := uint32(0)
		if comm.DiceCheck(smgp.Conf.SuccessRate) {
			// 失败消息的返回码
			rtCode = 39
		}
		resp := dly.ToResponse(rtCode).(*smgp.DeliverResp)
		// 发送响应
		err := c.AsyncWrite(resp.Encode(), func(c gnet.Conn) error {
			log.Debugf("[%-9s] >>> %s", "OnTraffic", resp)
			return nil
		})
		if err != nil {
			log.Errorf("[%-9s] DELIVERY_RESP ERROR: %v", "OnTraffic", err)
		}
	})
	return gnet.None
}

// 处理上行消息Resp
func handleDeliveryResp(s *Server, c gnet.Conn, header *smgp.MessageHeader) gnet.Action {
	// check connect
	_, ok := s.conMap.Load(c.RemoteAddr().String())
	if !ok {
		log.Warnf("[%-9s] unLogin connection: %s, closing...", "OnTraffic", c.RemoteAddr())
		return gnet.Close
	}
	frame := comm.TakeBytes(c, int(header.PacketLength-smgp.HeadLength))
	comm.LogHex(logging.DebugLevel, "Deliver", frame)

	resp := &smgp.DeliverResp{}
	err := resp.Decode(header, frame)
	if err != nil {
		log.Errorf("[%-9s] DELIVER_RESP ERROR: %v", "OnTraffic", err)
		return gnet.Close
	}
	log.Debugf("[%-9s] <<< %s", "OnTraffic", resp)

	return gnet.None
}

func handleSubmit(s *Server, c gnet.Conn, header *smgp.MessageHeader) gnet.Action {
	// check connect
	_, ok := s.conMap.Load(c.RemoteAddr().String())
	if !ok {
		log.Warnf("[%-9s] unLogin connection: %s, closing...", "OnTraffic", c.RemoteAddr())
		return gnet.Close
	}

	frame := comm.TakeBytes(c, int(header.PacketLength-smgp.HeadLength))
	comm.LogHex(logging.DebugLevel, "Submit", frame)
	sub := &smgp.Submit{}
	err := sub.Decode(header, frame)
	if err != nil {
		log.Errorf("[%-9s] SUBMIT ERROR: %v", "OnTraffic", err)
		return gnet.Close
	}
	log.Debugf("[%-9s] <<< %s", "OnTraffic", sub)
	// handle message async
	_ = s.pool.Submit(mtAsyncHandler(s, c, sub))
	return gnet.None
}

func mtAsyncHandler(s *Server, c gnet.Conn, sub *smgp.Submit) func() {
	return func() {
		// 采用通道控制消息收发速度,向通道发送信号
		s.window <- struct{}{}
		defer func() {
			// defer函数消费信号，确保每个消息的信号最终都会被消费
			<-s.window
		}()

		// 模拟消息处理耗时，可配置
		processTime := time.Duration(0)
		if smgp.Conf.MinSubmitRespMs > 0 && smgp.Conf.MaxSubmitRespMs > smgp.Conf.MinSubmitRespMs {
			processTime = time.Duration(comm.RandNum(smgp.Conf.MinSubmitRespMs, smgp.Conf.MaxSubmitRespMs))
			time.Sleep(processTime * time.Millisecond)
		}

		rtCode := uint32(0)
		if comm.DiceCheck(smgp.Conf.SuccessRate) {
			// 失败消息的返回码
			rtCode = 39
		}
		resp := sub.ToResponse(rtCode).(*smgp.SubmitResp)
		// 发送响应
		err := c.AsyncWrite(resp.Encode(), func(c gnet.Conn) error {
			log.Debugf("[%-9s] >>> %s", "OnTraffic", resp)
			return nil
		})
		if err != nil {
			log.Errorf("[%-9s] SUBMIT_RESP ERROR: %v", "OnTraffic", err)
		}

		// 发送状态报告
		if resp.Status() == 0 {
			_ = s.pool.Submit(reportAsyncSender(c, sub, resp.MsgId(), processTime))
		}
	}
}

func reportAsyncSender(c gnet.Conn, sub *smgp.Submit, msgId []byte, wait time.Duration) func() {
	return func() {
		if comm.DiceCheck(100) {
			return
		}
		dly := smgp.NewDeliveryReport(sub, msgId)
		// 模拟状态报告发送前的耗时
		if smgp.Conf.FixReportRespMs > 0 {
			processTime := wait + time.Duration(smgp.Conf.FixReportRespMs)
			time.Sleep(processTime * time.Millisecond)
		}
		// 发送状态报告
		err := c.AsyncWrite(dly.Encode(), func(c gnet.Conn) error {
			log.Debugf("[%-9s] >>> %s", "OnTraffic", dly)
			return nil
		})
		if err != nil {
			log.Errorf("[%-9s] DELIVERY_REPORT ERROR: %v", "OnTraffic", err)
		}
	}
}

func handActive(s *Server, c gnet.Conn, header *smgp.MessageHeader) (action gnet.Action) {
	resp := smgp.NewActiveTestResp(header.SequenceId)
	// send active_resp async
	_ = s.pool.Submit(func() {
		err := c.AsyncWrite(resp.Encode(), func(c gnet.Conn) error {
			log.Infof("[%-9s] >>> %s", "OnTraffic", resp)
			return nil
		})
		if err != nil {
			log.Errorf("[%-9s] ACTIVE_TEST_RESP ERROR: %v", "OnTraffic", err)
		}
	})
	return gnet.None
}

func handActiveResp(c gnet.Conn, header *smgp.MessageHeader) (action gnet.Action) {
	log.Infof("[%-9s] <<< %s from %s", "OnTraffic", header, c.RemoteAddr())
	return gnet.None
}

func getHeader(c gnet.Conn) *smgp.MessageHeader {
	frame := comm.TakeBytes(c, smgp.HeadLength)
	if frame == nil {
		return nil
	}
	comm.LogHex(logging.DebugLevel, "Header", frame)

	header := smgp.MessageHeader{}
	err := header.Decode(frame)
	if err != nil {
		log.Errorf("[%-9s] decode error: %v", "OnTraffic", err)
		return nil
	}
	return &header
}

func checkReceiveWindow(s *Server, c gnet.Conn, header *smgp.MessageHeader) gnet.Action {
	if len(s.window) == smgp.Conf.ReceiveWindowSize && header.RequestId == smgp.CmdSubmit {
		log.Warnf("[%-9s] FLOW CONTROL：receive window threshold reached.", "OnTraffic")
		l := int(header.PacketLength - smgp.HeadLength)
		discard, err := c.Discard(l)
		if err != nil || discard != l {
			return gnet.Close
		}
		sub := &smgp.Submit{}
		sub.MessageHeader = header
		resp := sub.ToResponse(1).(*smgp.SubmitResp)
		// 发送响应
		err = c.AsyncWrite(resp.Encode(), func(c gnet.Conn) error {
			log.Debugf("[%-9s] >>> %s", "OnTraffic", resp)
			return nil
		})
		if err != nil {
			log.Errorf("[%-9s] SUBMIT_RESP ERROR: %v", "OnTraffic", err)
			return gnet.Close
		}
		header.RequestId = 0
	}
	return gnet.None
}
