package cmpp

import (
	"encoding/binary"
	"fmt"
	"unsafe"
)

type MessageHeader struct {
	TotalLength uint32
	CommandId   uint32
	SequenceId  uint32
}

func (header *MessageHeader) Encode() []byte {
	if header.TotalLength < HEAD_LENGTH {
		header.TotalLength = HEAD_LENGTH
	}
	frame := make([]byte, header.TotalLength)
	binary.BigEndian.PutUint32(frame[0:4], header.TotalLength)
	binary.BigEndian.PutUint32(frame[4:8], header.CommandId)
	binary.BigEndian.PutUint32(frame[8:12], header.SequenceId)
	return frame
}

func (header *MessageHeader) Decode(frame []byte) error {
	if len(frame) < HEAD_LENGTH {
		return ErrorPacket
	}
	header.TotalLength = binary.BigEndian.Uint32(frame[0:4])
	header.CommandId = binary.BigEndian.Uint32(frame[4:8])
	header.SequenceId = binary.BigEndian.Uint32(frame[8:12])
	return nil
}

func (header *MessageHeader) String() string {
	return fmt.Sprintf("{ PacketLength: %d, RequestId: %s, SequenceId: %d }", header.TotalLength, CommandMap[header.CommandId], header.SequenceId)
}

func TrimStr(bts []byte) string {
	var i = 0
	for ; i < len(bts); i++ {
		if bts[i] == 0 {
			break
		}
	}
	ns := bts[:i]
	return *(*string)(unsafe.Pointer(&ns))
}

func V3() bool {
	return int(Conf.Version)&0xf0 == 0x30
}

const (
	HEAD_LENGTH           = 12                 // 报文头长度
	CMPP_CONNECT          = uint32(0x00000001) // 请求连接
	CMPP_CONNECT_RESP     = uint32(0x80000001) // 请求连接应答
	CMPP_TERMINATE        = uint32(0x00000002) // 终止连接
	CMPP_TERMINATE_RESP   = uint32(0x80000002) // 终止连接应答
	CMPP_SUBMIT           = uint32(0x00000004) // 提交短信
	CMPP_SUBMIT_RESP      = uint32(0x80000004) // 提交短信应答
	CMPP_DELIVER          = uint32(0x00000005) // 短信下发
	CMPP_DELIVER_RESP     = uint32(0x80000005) // 下发短信应答
	CMPP_ACTIVE_TEST      = uint32(0x00000008) // 激活测试
	CMPP_ACTIVE_TEST_RESP = uint32(0x80000008) // 激活测试应答
	// CMPP_QUERY                     = uint32(0x00000006) // 发送短信状态查询
	// CMPP_QUERY_RESP                = uint32(0x80000006) // 发送短信状态查询应答
	// CMPP_CANCEL                    = uint32(0x00000007) // 删除短信
	// CMPP_CANCEL_RESP               = uint32(0x80000007) // 删除短信应答
	// CMPP_FWD                       = uint32(0x00000009) // 消息前转
	// CMPP_FWD_RESP                  = uint32(0x80000009) // 消息前转应答
	// CMPP_MT_ROUTE                  = uint32(0x00000010) // MT 路由请求
	// CMPP_MT_ROUTE_RESP             = uint32(0x80000010) // MT 路由请求应答
	// CMPP_MO_ROUTE                  = uint32(0x00000011) // MO 路由请求
	// CMPP_MO_ROUTE_RESP             = uint32(0x80000011) // MO 路由请求应答
	// CMPP_GET_MT_ROUTE              = uint32(0x00000012) // 获取 MT 路由请求
	// CMPP_GET_MT_ROUTE_RESP         = uint32(0x80000012) // 获取 MT 路由请求应答
	// CMPP_MT_ROUTE_UPDATE           = uint32(0x00000013) // MT 路由更新
	// CMPP_MT_ROUTE_UPDATE_RESP      = uint32(0x80000013) // MT 路由更新应答
	// CMPP_MO_ROUTE_UPDATE           = uint32(0x00000014) // MO 路由更新
	// CMPP_MO_ROUTE_UPDATE_RESP      = uint32(0x80000014) // MO 路由更新应答
	// CMPP_PUSH_MT_ROUTE_UPDATE      = uint32(0x00000015) // MT 路由更新
	// CMPP_PUSH_MT_ROUTE_UPDATE_RESP = uint32(0x80000015) // MT 路由更新应答
	// CMPP_PUSH_MO_ROUTE_UPDATE      = uint32(0x00000016) // MO 路由更新
	// CMPP_PUSH_MO_ROUTE_UPDATE_RESP = uint32(0x80000016) // MO 路由更新应答
	// CMPP_GET_MO_ROUTE              = uint32(0x00000017) // 获取 MO 路由请求
	// CMPP_GET_MO_ROUTE_RESP         = uint32(0x80000017) // 获取 MO 路由请求应答
)

var CommandMap = make(map[uint32]string)

func init() {
	CommandMap[CMPP_CONNECT] = "CMPP_CONNECT"
	CommandMap[CMPP_CONNECT_RESP] = "CMPP_CONNECT_RESP"
	CommandMap[CMPP_TERMINATE] = "CMPP_TERMINATE"
	CommandMap[CMPP_TERMINATE_RESP] = "CMPP_TERMINATE_RESP"
	CommandMap[CMPP_SUBMIT] = "CMPP_SUBMIT"
	CommandMap[CMPP_SUBMIT_RESP] = "CMPP_SUBMIT_RESP"
	CommandMap[CMPP_DELIVER] = "CMPP_DELIVER"
	CommandMap[CMPP_DELIVER_RESP] = "CMPP_DELIVER_RESP"
	CommandMap[CMPP_ACTIVE_TEST] = "CMPP_ACTIVE_TEST"
	CommandMap[CMPP_ACTIVE_TEST_RESP] = "CMPP_ACTIVE_TEST_RESP"
}
