package cmcc

import (
	"encoding/binary"
	"fmt"

	"sms-vgateway/snowflake32"
)

var SmscSequence = snowflake32.NewSnowflake(1, 1)

type Report struct {
	msgId          uint64 // 信息标识，SP提交短信(CMPP_SUBMIT)操作时，与SP相连的ISMG产生的 Msg_Id。【8字节】
	stat           string // 发送短信的应答结果。【7字节】
	submitTime     string // yyMMddHHmm 【10字节】
	doneTime       string // yyMMddHHmm 【10字节】
	destTerminalId string // SP 发送 CMPP_SUBMIT 消息的目标终端  【21字节】
	smscSequence   uint32 // 取自SMSC发送状态报告的消息体中的消息标识。【4字节】
}

func (rt *Report) Encode() []byte {
	frame := make([]byte, 60)
	binary.BigEndian.PutUint64(frame[0:8], rt.msgId)
	copy(frame[8:15], rt.stat)
	copy(frame[15:25], rt.submitTime)
	copy(frame[25:35], rt.doneTime)
	copy(frame[35:56], rt.destTerminalId)
	binary.BigEndian.PutUint32(frame[56:60], rt.smscSequence)
	return frame
}

func (rt *Report) Decode(frame []byte) error {
	if len(frame) < 60 {
		return ErrorPacket
	}
	rt.msgId = binary.BigEndian.Uint64(frame[0:8])
	rt.stat = TrimStr(frame[8:15])
	rt.submitTime = TrimStr(frame[15:25])
	rt.doneTime = TrimStr(frame[25:35])
	rt.destTerminalId = TrimStr(frame[35:56])
	rt.smscSequence = binary.BigEndian.Uint32(frame[56:60])
	return nil
}

func (rt *Report) String() string {
	return fmt.Sprintf("{ msgId: %d, stat: %s, submitTime: %s, doneTime: %s, destTerminalId: %s, smscSequence: %d }",
		rt.msgId, rt.stat, rt.submitTime, rt.doneTime, rt.destTerminalId, rt.smscSequence)
}
