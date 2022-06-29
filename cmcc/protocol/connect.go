package cmcc

import (
	"bytes"
	"crypto/md5"
	"encoding/binary"
	"fmt"
	"math/rand"
	"strconv"
	"time"
)

type CmppConnect struct {
	*MessageHeader             // +12 = 12
	SourceAddr          string // +6 = 18
	AuthenticatorSource string // +16 = 34
	Version             uint8  // +1 = 35
	Timestamp           uint32 // +4 = 39
}

func NewConnect() *CmppConnect {
	con := &CmppConnect{}
	header := &MessageHeader{}
	header.TotalLength = 39
	header.CommandId = CMPP_CONNECT
	header.SequenceId = rand.Uint32()
	con.MessageHeader = header
	con.Version = Conf.Version
	con.SourceAddr = Conf.SourceAddr
	// 2006-1-2 15:04:05
	ts, _ := strconv.ParseUint(time.Now().Format("0102150405"), 10, 32)
	con.Timestamp = uint32(ts)
	ss := reqAuthMd5(con)
	con.AuthenticatorSource = string(ss[:])
	return con
}

func (connect *CmppConnect) Encode() []byte {
	frame := connect.MessageHeader.Encode()
	if len(frame) == LEN_CMPP_CONNECT && connect.TotalLength == LEN_CMPP_CONNECT {
		copy(frame[12:18], connect.SourceAddr)
		copy(frame[18:34], connect.AuthenticatorSource)
		frame[34] = connect.Version
		binary.BigEndian.PutUint32(frame[35:39], connect.Timestamp)
	}
	return frame
}

func (connect *CmppConnect) Decode(header *MessageHeader, frame []byte) error {
	// check
	if header == nil || header.CommandId != CMPP_CONNECT || len(frame) < (LEN_CMPP_CONNECT-HEAD_LENGTH) {
		return ErrorPacket
	}
	connect.MessageHeader = header
	connect.SourceAddr = string(frame[0:6])
	connect.AuthenticatorSource = string(frame[6:22])
	connect.Version = frame[22]
	connect.Timestamp = binary.BigEndian.Uint32(frame[23:27])
	return nil
}

func (connect *CmppConnect) String() string {
	return fmt.Sprintf("{ Header: %s, SourceAddr: %s, AuthenticatorSource: %x, Version: %x, Timestamp: %v }",
		connect.MessageHeader, connect.SourceAddr, connect.AuthenticatorSource, connect.Version, connect.Timestamp)
}

func (connect *CmppConnect) Check() uint32 {
	if connect.Version > Conf.Version {
		return 4
	}

	authSource := []byte(connect.AuthenticatorSource)
	authMd5 := reqAuthMd5(connect)
	i := bytes.Compare(authSource, authMd5[:])
	if i == 0 {
		return 0
	}
	return 3
}

func (connect *CmppConnect) ToResponse() *CmppConnectResp {
	response := &CmppConnectResp{}
	header := &MessageHeader{}
	header.TotalLength = LEN_CMPP_CONNECT_RESP
	header.CommandId = CMPP_CONNECT_RESP
	header.SequenceId = connect.SequenceId
	response.MessageHeader = header
	response.Status = connect.Check()
	// AuthenticatorISMG =MD5 ( Status+AuthenticatorSource+shar ed secret)
	authDt := make([]byte, 0, 64)
	authDt = append(authDt, fmt.Sprintf("%d", response.Status)...)
	authDt = append(authDt, connect.AuthenticatorSource...)
	authDt = append(authDt, Conf.SharedSecret...)
	auth := md5.Sum(authDt)
	response.AuthenticatorISMG = string(auth[:])
	response.Version = Conf.Version
	return response
}

func reqAuthMd5(connect *CmppConnect) [16]byte {
	// AuthenticatorSource = MD5(Source_Addr+9 字节的 0 +shared secret+timestamp)
	// timestamp 格式为: MMDDHHMMSS，即月日时分秒，10 位。
	authDt := make([]byte, 0, 64)
	authDt = append(authDt, Conf.SourceAddr...)
	authDt = append(authDt, 0, 0, 0, 0, 0, 0, 0, 0, 0)
	authDt = append(authDt, Conf.SharedSecret...)
	authDt = append(authDt, fmt.Sprintf("%d", connect.Timestamp)...)
	authMd5 := md5.Sum(authDt)
	return authMd5
}

type CmppConnectResp struct {
	*MessageHeader           // +12 = 12
	Status            uint32 // +4 = 16
	AuthenticatorISMG string // +16 = 32
	Version           uint8  // +1 = 33
}

func (resp *CmppConnectResp) Encode() []byte {
	frame := resp.MessageHeader.Encode()
	if len(frame) == LEN_CMPP_CONNECT_RESP && resp.TotalLength == LEN_CMPP_CONNECT_RESP {
		binary.BigEndian.PutUint32(frame[12:16], resp.Status)
		copy(frame[16:32], resp.AuthenticatorISMG)
		frame[32] = resp.Version
	}
	return frame
}

func (resp *CmppConnectResp) Decode(header *MessageHeader, frame []byte) error {
	// check
	if header == nil || header.CommandId != CMPP_CONNECT_RESP || len(frame) < (LEN_CMPP_CONNECT_RESP-HEAD_LENGTH) {
		return ErrorPacket
	}
	resp.MessageHeader = header
	resp.Status = binary.BigEndian.Uint32(frame[0:4])
	resp.AuthenticatorISMG = string(frame[4:20])
	resp.Version = frame[20]
	return nil
}

func (resp *CmppConnectResp) String() string {
	return fmt.Sprintf("{ Header: %s, Status: %v, AuthenticatorISMG: %x, Version: %x }",
		resp.MessageHeader, resp.Status, resp.AuthenticatorISMG, resp.Version)
}

var ConnectStatusMap = make(map[uint32]string)

func init() {
	ConnectStatusMap[0] = "成功"
	ConnectStatusMap[1] = "消息结构错"
	ConnectStatusMap[2] = "非法源地址"
	ConnectStatusMap[3] = "认证错"
	ConnectStatusMap[4] = "版本太高"
	ConnectStatusMap[5] = "其他错误"
}
