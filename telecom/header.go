package telecom

import (
	"encoding/binary"
	"fmt"
)

type MessageHeader struct {
	PacketLength uint32
	RequestId    uint32
	SequenceId   uint32
}

func (header *MessageHeader) Encode() []byte {
	if header.PacketLength < HeadLength {
		header.PacketLength = HeadLength
	}
	frame := make([]byte, header.PacketLength)
	binary.BigEndian.PutUint32(frame[0:4], header.PacketLength)
	binary.BigEndian.PutUint32(frame[4:8], header.RequestId)
	binary.BigEndian.PutUint32(frame[8:12], header.SequenceId)
	return frame
}

func (header *MessageHeader) Decode(frame []byte) error {
	if len(frame) < HeadLength {
		return ErrorPacket
	}
	header.PacketLength = binary.BigEndian.Uint32(frame[0:4])
	header.RequestId = binary.BigEndian.Uint32(frame[4:8])
	header.SequenceId = binary.BigEndian.Uint32(frame[8:12])
	return nil
}

func (header *MessageHeader) String() string {
	return fmt.Sprintf("{ PacketLength: %d, RequestId: %s, SequenceId: %d }", header.PacketLength, CommandMap[header.RequestId], header.SequenceId)
}

const (
	HeadLength        = 12         // 报文头长度
	CmdLogin          = 0x00000001 // 客户端登录
	CmdLoginResp      = 0x80000001 // 客户端登录应答
	CmdSubmit         = 0x00000002 // 提交短消息
	CmdSubmitResp     = 0x80000002 // 提交短消息应答
	CmdDeliver        = 0x00000003 // 下发短消息
	CmdDeliverResp    = 0x80000003 // 下发短消息应答
	CmdActiveTest     = 0x00000004 // 链路检测
	CmdActiveTestResp = 0x80000004 // 链路检测应答
	CmdExit           = 0x00000006 // 退出请求
	CmdExitResp       = 0x80000006 // 退出应答
	// Forward                   = 0x00000005 // 短消息前转
	// Forward_Resp              = 0x80000005 // 短消息前转应答
	// Query                     = 0x00000007 // SP统计查询
	// Query_Resp                = 0x80000007 // SP统计查询应答
	// Query_TE_Route            = 0x00000008 // 查询TE路由
	// Query_TE_Route_Resp       = 0x80000008 // 查询TE路由应答
	// Query_SP_Route            = 0x00000009 // 查询SP路由
	// Query_SP_Route_Resp       = 0x80000009 // 查询SP路由应答
	// Payment_Request           = 0x0000000A // 扣款请求(用于预付费系统，参见增值业务计费方案)
	// Payment_Request_Resp      = 0x8000000A // 扣款请求响应(用于预付费系统，参见增值业务计费方案，下同)
	// Payment_Affirm            = 0x0000000B // 扣款确认(用于预付费系统，参见增值业务计费方案)
	// Payment_Affirm_Resp       = 0x8000000B // 扣款确认响应(用于预付费系统，参见增值业务计费方案)
	// Query_UserState           = 0x0000000C // 查询用户状态(用于预付费系统，参见增值业务计费方案)
	// Query_UserState_Resp      = 0x8000000C // 查询用户状态响应(用于预付费系统，参见增值业务计费方案)
	// Get_All_TE_Route          = 0x0000000D // 获取所有终端路由
	// Get_All_TE_Route_Resp     = 0x8000000D // 获取所有终端路由应答
	// Get_All_SP_Route          = 0x0000000E // 获取所有SP路由
	// Get_All_SP_Route_Resp     = 0x8000000E // 获取所有SP路由应答
	// Update_TE_Route           = 0x0000000F // SMGW向GNS更新终端路由
	// Update_TE_Route_Resp      = 0x8000000F // SMGW向GNS更新终端路由应答
	// Update_SP_Route           = 0x00000010 // SMGW向GNS更新SP路由
	// Update_SP_Route_Resp      = 0x80000010 // SMGW向GNS更新SP路由应答
	// Push_Update_TE_Route      = 0x00000011 // GNS向SMGW更新终端路由
	// Push_Update_TE_Route_Resp = 0x80000011 // GNS向SMGW更新终端路由应答
	// Push_Update_SP_Route      = 0x00000012 // GNS向SMGW更新SP路由
	// Push_Update_SP_Route_Resp = 0x80000012 // GNS向SMGW更新SP路由应答
)

var CommandMap = make(map[uint32]string)

func init() {
	CommandMap[CmdLogin] = "Login"
	CommandMap[CmdLoginResp] = "Login_Resp"
	CommandMap[CmdSubmit] = "Submit"
	CommandMap[CmdSubmitResp] = "Submit_Resp"
	CommandMap[CmdDeliver] = "Deliver"
	CommandMap[CmdDeliverResp] = "Deliver_Resp"
	CommandMap[CmdActiveTest] = "Active_Test"
	CommandMap[CmdActiveTestResp] = "Active_Test_Resp"
	CommandMap[CmdExit] = "Exit"
	CommandMap[CmdExitResp] = "Exit_Resp"
	// CommandMap[Forward] = "Forward"
	// CommandMap[Forward_Resp] = "Forward_Resp"
	// CommandMap[Query] = "Query"
	// CommandMap[Query_Resp] = "Query_Resp"
	// CommandMap[Query_TE_Route] = "Query_TE_Route"
	// CommandMap[Query_TE_Route_Resp] = "Query_TE_Route_Resp"
	// CommandMap[Query_SP_Route] = "Query_SP_Route"
	// CommandMap[Query_SP_Route_Resp] = "Query_SP_Route_Resp"
	// CommandMap[Payment_Request] = "Payment_Request"
	// CommandMap[Payment_Request_Resp] = "Payment_Request_Resp"
	// CommandMap[Payment_Affirm] = "Payment_Affirm"
	// CommandMap[Payment_Affirm_Resp] = "Payment_Affirm_Resp"
	// CommandMap[Query_UserState] = "Query_UserState"
	// CommandMap[Query_UserState_Resp] = "Query_UserState_Resp"
	// CommandMap[Get_All_TE_Route] = "Get_All_TE_Route"
	// CommandMap[Get_All_TE_Route_Resp] = "Get_All_TE_Route_Resp"
	// CommandMap[Get_All_SP_Route] = "Get_All_SP_Route"
	// CommandMap[Get_All_SP_Route_Resp] = "Get_All_SP_Route_Resp"
	// CommandMap[Update_TE_Route] = "Update_TE_Route"
	// CommandMap[Update_TE_Route_Resp] = "Update_TE_Route_Resp"
	// CommandMap[Update_SP_Route] = "Update_SP_Route"
	// CommandMap[Update_SP_Route_Resp] = "Update_SP_Route_Resp"
	// CommandMap[Push_Update_TE_Route] = "Push_Update_TE_Route"
	// CommandMap[Push_Update_TE_Route_Resp] = "Push_Update_TE_Route_Resp"
	// CommandMap[Push_Update_SP_Route] = "Push_Update_SP_Route"
	// CommandMap[Push_Update_SP_Route_Resp] = "Push_Update_SP_Route_Resp"
}
