package main

import (
	"flag"
	"fmt"
	"sync"
	"time"

	"github.com/panjf2000/gnet/v2"
	"github.com/panjf2000/gnet/v2/pkg/logging"
	"github.com/panjf2000/gnet/v2/pkg/pool/goroutine"

	cmcc "sms-vgateway/cmcc/protocol"
)

//
// export GNET_LOGGING_FILE="/Users/huangzhonghui/logs/sms.log"
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
		protocol:  "tcp",
		address:   fmt.Sprintf(":%d", port),
		multicore: multicore,
		pool:      pool,
	}
	err := gnet.Run(ss, ss.protocol+"://"+ss.address, gnet.WithMulticore(multicore))
	log.Infof("server exits with error: %v", err)
}

type Server struct {
	gnet.BuiltinEventEngine
	engine           gnet.Engine
	protocol         string
	address          string
	multicore        bool
	pool             *goroutine.Pool
	connectedSockets sync.Map
}

func (s *Server) OnBoot(eng gnet.Engine) (action gnet.Action) {
	log.Infof("[%-10s] running server on %s with multi-core=%t", "OnBoot", fmt.Sprintf("%s://%s", s.protocol, s.address), s.multicore)
	s.engine = eng
	return
}

func (s *Server) OnShutdown(eng gnet.Engine) {
	log.Infof("[%-10s] shutdown server %s ...", "OnShutdown", fmt.Sprintf("%s://%s", s.protocol, s.address))
	for eng.CountConnections() > 0 {
		log.Infof("[%-10s] active connections is %d, waiting...", eng.CountConnections())
		time.Sleep(10 * time.Millisecond)
	}
	log.Infof("[%-10s] shutdown server %s completed!", "OnShutdown", fmt.Sprintf("%s://%s", s.protocol, s.address))
}

func (s *Server) OnOpen(c gnet.Conn) (out []byte, action gnet.Action) {
	log.Infof("[%-10s] [%v<->%v] activeCons=%d.", "OnOpen", c.RemoteAddr(), c.LocalAddr(), s.activeCons())
	return
}

func (s *Server) OnClose(c gnet.Conn, e error) (action gnet.Action) {
	log.Warnf("[%-10s] [%v<->%v] activeCons=%d, reason=%+v.", "OnClose", c.RemoteAddr(), c.LocalAddr(), s.activeCons(), e)
	return
}

// OnTraffic fires when a local socket receives data from the peer.
func (s *Server) OnTraffic(c gnet.Conn) (action gnet.Action) {
	header := getHeader(c)
	if header == nil {
		log.Warnf("[%-10s] [%v<->%v] decode error, close session...", "OnTraffic", c.RemoteAddr(), c.LocalAddr())
		return gnet.Close
	}

	switch header.CommandId {
	case cmcc.CMPP_CONNECT:
		return handleConnect(s, c, header)
	case cmcc.CMPP_CONNECT_RESP:
	case cmcc.CMPP_SUBMIT:
	case cmcc.CMPP_SUBMIT_RESP:
	case cmcc.CMPP_DELIVER:
	case cmcc.CMPP_DELIVER_RESP:
	case cmcc.CMPP_ACTIVE_TEST:
	case cmcc.CMPP_ACTIVE_TEST_RESP:
	case cmcc.CMPP_TERMINATE:
	case cmcc.CMPP_TERMINATE_RESP:
	}

	return gnet.None
}

// OnTick fires immediately after the engine starts and will fire again
// following the duration specified by the delay return value.
func (s *Server) OnTick() (delay time.Duration, action gnet.Action) {
	return
}

func handleConnect(s *Server, c gnet.Conn, header *cmcc.MessageHeader) (action gnet.Action) {
	action = gnet.None
	frame := take(c, cmcc.LEN_CMPP_CONNECT-cmcc.HEAD_LENGTH)
	log.Debugf("%-10s Hex: %x", "Connect", frame)

	connect := &cmcc.CmppConnect{}
	err := connect.Decode(header, frame)
	if err != nil {
		log.Errorf("CMPP_CONNECT ERROR: %v", err)
		return gnet.Close
	}

	log.Infof("%-9s<<< %s", "receive", connect)
	resp := connect.ToResponse()
	if resp.Status != 0 {
		log.Errorf("CMPP_CONNECT ERROR: Auth Error, Status=(%d,%s)", resp.Status, cmcc.ConnectStatusMap[resp.Status])
		action = gnet.Close
	}
	if s.activeCons() >= cmcc.Conf.MaxCons {
		resp.Status = 5 // 其他错误,连接数过多
		action = gnet.Close
	}

	// send cmpp_connect_resp
	_ = s.pool.Submit(func() {
		err := c.AsyncWrite(resp.Encode(), func(c gnet.Conn) error {
			s.connectedSockets.Store(c.RemoteAddr().String(), c)
			log.Infof("%-9s>>> %s", "send", resp)
			return nil
		})
		if err != nil {
			log.Errorf("CMPP_CONNECT_RESP ERROR: %v", err)
		}
	})
	return action
}

func getHeader(c gnet.Conn) *cmcc.MessageHeader {
	frame := take(c, cmcc.HEAD_LENGTH)
	if frame == nil {
		return nil
	}
	log.Debugf("%-10s Hex: %x", "Header", frame)
	header := cmcc.MessageHeader{}
	err := header.Decode(frame)
	if err != nil {
		log.Errorf("decode error: %v", err)
		return nil
	}
	return &header
}

// 消费一定字节数的数据
func take(c gnet.Conn, bytes int) []byte {
	if c.InboundBuffered() < bytes {
		return nil
	}
	frame, err := c.Peek(bytes)
	if err != nil {
		log.Errorf("decode error: %v", err)
		return nil
	}
	_, err = c.Discard(bytes)
	if err != nil {
		log.Errorf("decode error: %v", err)
		return nil
	}
	return frame
}

func (s *Server) activeCons() int {
	return s.engine.CountConnections()
}
