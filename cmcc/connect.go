package cmcc

import (
	"bytes"
	"crypto/md5"
	"encoding/binary"
	"fmt"
	"strconv"
	"time"
)

type Connect struct {
	*MessageHeader             // +12 = 12：消息头
	sourceAddr          string // +6 = 18：源地址，此处为 SP_Id
	authenticatorSource string // +16 = 34： 用于鉴别源地址。其值通过单向 MD5 hash 计算得出，表示如下: authenticatorSource = MD5(Source_Addr+9 字节的 0 +shared secret+timestamp) Shared secret 由中国移动与源地址实 体事先商定，timestamp 格式为: MMDDHHMMSS，即月日时分秒，10 位。
	version             uint8  // +1 = 35：双方协商的版本号(高位 4bit 表示主 版本号,低位 4bit 表示次版本号)，对 于3.0的版本，高4bit为3，低4位为 0
	timestamp           uint32 // +4 = 39：时间戳的明文,由客户端产生,格式为 MMDDHHMMSS，即月日时分秒，10 位数字的整型，右对齐。
}

func NewConnect() *Connect {
	con := &Connect{}
	header := &MessageHeader{}
	header.TotalLength = 39
	header.CommandId = CMPP_CONNECT
	header.SequenceId = uint32(Sequence32.NextVal())
	con.MessageHeader = header
	con.version = Conf.Version
	con.sourceAddr = Conf.SourceAddr
	// 2006-1-2 15:04:05
	ts, _ := strconv.ParseUint(time.Now().Format("0102150405"), 10, 32)
	con.timestamp = uint32(ts)
	ss := reqAuthMd5(con)
	con.authenticatorSource = string(ss[:])
	return con
}

func (connect *Connect) Encode() []byte {
	frame := connect.MessageHeader.Encode()
	if len(frame) == 39 && connect.TotalLength == 39 {
		copy(frame[12:18], connect.sourceAddr)
		copy(frame[18:34], connect.authenticatorSource)
		frame[34] = connect.version
		binary.BigEndian.PutUint32(frame[35:39], connect.timestamp)
	}
	return frame
}

func (connect *Connect) Decode(header *MessageHeader, frame []byte) error {
	// check
	if header == nil || header.CommandId != CMPP_CONNECT || len(frame) < (39-HEAD_LENGTH) {
		return ErrorPacket
	}
	connect.MessageHeader = header
	connect.sourceAddr = string(frame[0:6])
	connect.authenticatorSource = string(frame[6:22])
	connect.version = frame[22]
	connect.timestamp = binary.BigEndian.Uint32(frame[23:27])
	return nil
}

func (connect *Connect) String() string {
	return fmt.Sprintf("{ Header: %s, sourceAddr: %s, authenticatorSource: %x, version: %x, timestamp: %v }",
		connect.MessageHeader, connect.sourceAddr, connect.authenticatorSource, connect.version, connect.timestamp)
}

func (connect *Connect) Check() uint32 {
	if connect.version > Conf.Version {
		return 4
	}

	authSource := []byte(connect.authenticatorSource)
	authMd5 := reqAuthMd5(connect)
	i := bytes.Compare(authSource, authMd5[:])
	if i == 0 {
		return 0
	}
	return 3
}

func (connect *Connect) ToResponse(code uint32) interface{} {
	response := &ConnectResp{}
	header := &MessageHeader{}
	header.TotalLength = 33
	header.CommandId = CMPP_CONNECT_RESP
	header.SequenceId = connect.SequenceId
	response.MessageHeader = header
	if code == 0 {
		response.status = connect.Check()
	} else {
		response.status = code
	}
	// authenticatorISMG =MD5 ( status+authenticatorSource+shar ed secret)
	authDt := make([]byte, 0, 64)
	authDt = append(authDt, fmt.Sprintf("%d", response.status)...)
	authDt = append(authDt, connect.authenticatorSource...)
	authDt = append(authDt, Conf.SharedSecret...)
	auth := md5.Sum(authDt)
	response.authenticatorISMG = string(auth[:])
	response.version = Conf.Version
	return response
}

func reqAuthMd5(connect *Connect) [16]byte {
	// authenticatorSource = MD5(Source_Addr+9 字节的 0 +shared secret+timestamp)
	// timestamp 格式为: MMDDHHMMSS，即月日时分秒，10 位。
	authDt := make([]byte, 0, 64)
	authDt = append(authDt, Conf.SourceAddr...)
	authDt = append(authDt, 0, 0, 0, 0, 0, 0, 0, 0, 0)
	authDt = append(authDt, Conf.SharedSecret...)
	authDt = append(authDt, fmt.Sprintf("%d", connect.timestamp)...)
	authMd5 := md5.Sum(authDt)
	return authMd5
}

type ConnectResp struct {
	*MessageHeader           // +12 = 12
	status            uint32 // +4 = 16
	authenticatorISMG string // +16 = 32
	version           uint8  // +1 = 33
}

func (resp *ConnectResp) Encode() []byte {
	frame := resp.MessageHeader.Encode()
	if len(frame) == 33 && resp.TotalLength == 33 {
		binary.BigEndian.PutUint32(frame[12:16], resp.status)
		copy(frame[16:32], resp.authenticatorISMG)
		frame[32] = resp.version
	}
	return frame
}

func (resp *ConnectResp) Decode(header *MessageHeader, frame []byte) error {
	// check
	if header == nil || header.CommandId != CMPP_CONNECT_RESP || len(frame) < (33-HEAD_LENGTH) {
		return ErrorPacket
	}
	resp.MessageHeader = header
	resp.status = binary.BigEndian.Uint32(frame[0:4])
	resp.authenticatorISMG = string(frame[4:20])
	resp.version = frame[20]
	return nil
}

func (resp *ConnectResp) String() string {
	return fmt.Sprintf("{ Header: %s, status: {%d: %s}, authenticatorISMG: %x, version: %x }",
		resp.MessageHeader, resp.status, ConnectStatusMap[resp.status], resp.authenticatorISMG, resp.version)
}

func (resp *ConnectResp) Status() uint32 {
	return resp.status
}

var ConnectStatusMap = map[uint32]string{
	0: "成功",
	1: "消息结构错",
	2: "非法源地址",
	3: "认证错",
	4: "版本太高",
	5: "其他错误",
}
