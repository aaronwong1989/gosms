package smgp

import (
	"bytes"
	"crypto/md5"
	"encoding/binary"
	"fmt"
	"strconv"
	"time"

	"github.com/aaronwong1989/gosms/codec"
)

type Login struct {
	*MessageHeader             //  【12字节】消息头
	clientID            string //  【8字节】客户端用来登录服务器端的用户账号。
	authenticatorClient []byte //  【16字节】客户端认证码，用来鉴别客户端的合法性。
	loginMode           byte   //  【1字节】客户端用来登录服务器端的登录类型。
	timestamp           uint32 //  【4字节】时间戳
	version             byte   //  【1字节】客户端支持的协议版本号
}
type LoginResp struct {
	*MessageHeader             // 协议头, 12字节
	status              uint32 // 状态码，4字节
	authenticatorServer []byte // 认证串，16字节
	version             byte   // 版本，1字节
}

const (
	LoginLen     = 42
	LoginRespLen = 33
)

func NewLogin() *Login {
	lo := &Login{}
	header := &MessageHeader{}
	header.PacketLength = LoginLen
	header.RequestId = CmdLogin
	header.SequenceId = uint32(Seq32.NextVal())
	lo.MessageHeader = header
	lo.clientID = Conf.GetString("client-id")
	lo.loginMode = 2
	ts, _ := strconv.ParseUint(time.Now().Format("0102150405"), 10, 32)
	lo.timestamp = uint32(ts)
	// TODO TEST ONLY
	// lo.timestamp = uint32(705192634)
	ss := reqAuthMd5(lo)
	lo.authenticatorClient = ss[:]
	lo.version = byte(Conf.GetInt("version"))
	return lo
}

func (lo *Login) Encode() []byte {
	frame := lo.MessageHeader.Encode()
	if len(frame) == LoginLen && lo.PacketLength == LoginLen {
		copy(frame[12:20], lo.clientID)
		copy(frame[20:36], lo.authenticatorClient)
		frame[36] = lo.loginMode
		binary.BigEndian.PutUint32(frame[37:41], lo.timestamp)
		frame[41] = lo.version
	}
	return frame
}

func (lo *Login) Decode(header codec.IHead, frame []byte) error {
	h := header.(*MessageHeader)
	// check
	if header == nil || h.RequestId != CmdLogin || len(frame) < (LoginLen-HeadLength) {
		return ErrorPacket
	}
	lo.MessageHeader = h
	lo.clientID = string(frame[0:8])
	lo.authenticatorClient = frame[8:24]
	lo.loginMode = frame[24]
	lo.timestamp = binary.BigEndian.Uint32(frame[25:29])
	lo.version = frame[29]
	return nil
}

func (lo *Login) String() string {
	return fmt.Sprintf("{ Header: %s, clientID: %s, authenticatorClient: %x, logoinMode: %x, timestamp: %010d, version: %#x }",
		lo.MessageHeader, lo.clientID, lo.authenticatorClient, lo.loginMode, lo.timestamp, lo.version)
}

func (lo *Login) Check() uint32 {
	// 大版本不匹配
	if lo.version&0xf0 != byte(Conf.GetInt("version"))&0xf0 {
		return 22
	}

	authSource := lo.authenticatorClient
	authMd5 := reqAuthMd5(lo)
	log.Debugf("[AuthCheck] input  : %x", authSource)
	log.Debugf("[AuthCheck] compute: %x", authMd5)
	i := bytes.Compare(authSource, authMd5[:])
	// 配置不做校验或校验通过时返回0
	if !Conf.GetBool("auth-check") || i == 0 {
		return 0
	}
	return 21
}

func (lo *Login) ToResponse(code uint32) interface{} {
	response := &LoginResp{}
	header := &MessageHeader{}
	header.PacketLength = LoginRespLen
	header.RequestId = CmdLoginResp
	header.SequenceId = lo.SequenceId
	response.MessageHeader = header
	if code == 0 {
		response.status = lo.Check()
	} else {
		response.status = code
	}
	authDt := make([]byte, 0, 64)
	authDt = append(authDt, fmt.Sprintf("%d", response.status)...)
	authDt = append(authDt, lo.authenticatorClient...)
	authDt = append(authDt, Conf.GetString("shared-secret")...)
	auth := md5.Sum(authDt)
	response.authenticatorServer = auth[:]
	response.version = byte(Conf.GetInt("version"))
	return response
}

func reqAuthMd5(connect *Login) [16]byte {
	authDt := make([]byte, 0, 64)
	authDt = append(authDt, Conf.GetString("client-id")...)
	authDt = append(authDt, 0, 0, 0, 0, 0, 0, 0)
	authDt = append(authDt, Conf.GetString("shared-secret")...)
	authDt = append(authDt, fmt.Sprintf("%010d", connect.timestamp)...)
	log.Debugf("[AuthCheck] auth data: %x", authDt)
	authMd5 := md5.Sum(authDt)
	return authMd5
}

func (resp *LoginResp) Encode() []byte {
	frame := resp.MessageHeader.Encode()
	var index int
	if len(frame) == int(resp.PacketLength) {
		index = 12
		binary.BigEndian.PutUint32(frame[index:index+4], resp.status)
		index += 4
		copy(frame[index:index+16], resp.authenticatorServer)
		index += 16
		frame[index] = resp.version
	}
	return frame
}

func (resp *LoginResp) Decode(header codec.IHead, frame []byte) error {
	h := header.(*MessageHeader)
	// check
	if h == nil || h.RequestId != CmdLoginResp || len(frame) < (LoginRespLen-HeadLength) {
		return ErrorPacket
	}
	var index int
	resp.MessageHeader = h
	resp.status = binary.BigEndian.Uint32(frame[0 : index+4])
	index = 4
	resp.authenticatorServer = frame[index : index+16]
	index += 16
	resp.version = frame[index]
	return nil
}

func (resp *LoginResp) String() string {
	return fmt.Sprintf("{ Header: %s, status: {%d: %s}, authenticatorISMG: %x, version: %#x }",
		resp.MessageHeader, resp.status, ConnectStatusMap[resp.status], resp.authenticatorServer, resp.version)
}

func (resp *LoginResp) Status() uint32 {
	return resp.status
}

var ConnectStatusMap = map[uint32]string{
	0:  "成功",
	1:  "系统忙",
	2:  "超过最大连接数",
	10: "消息结构错",
	11: "命令字错",
	12: "序列号重复",
	20: "IP地址错",
	21: "认证错",
	22: "版本太高",
	30: "非法消息类型（MsgType）",
	31: "非法优先级（Priority）",
	32: "非法资费类型（FeeType）",
	33: "非法资费代码（FeeCode）",
}
