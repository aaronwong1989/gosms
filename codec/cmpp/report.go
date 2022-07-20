package cmpp

import (
	"encoding/binary"
	"fmt"

	"github.com/aaronwong1989/gosms/codec"
)

var ReportSeq codec.Sequence32

type Report struct {
	msgId          uint64 // 信息标识，SP提交短信(CMPP_SUBMIT)操作时，与SP相连的ISMG产生的 Msg_Id。【8字节】
	stat           string // 发送短信的应答结果。【7字节】
	submitTime     string // yyMMddHHmm 【10字节】
	doneTime       string // yyMMddHHmm 【10字节】
	destTerminalId string // SP 发送 CMPP_SUBMIT 消息的目标终端  【21字节】
	smscSequence   uint32 // 取自SMSC发送状态报告的消息体中的消息标识。【4字节】
}

func NewReport(msgId uint64, destTerminalId string, submitTime string, doneTime string) *Report {
	report := &Report{msgId: msgId, submitTime: submitTime, doneTime: doneTime, destTerminalId: destTerminalId}
	report.smscSequence = uint32(ReportSeq.NextVal())
	// 判断序号的时间戳部分
	switch (report.smscSequence >> 14) % 100 {
	case 99:
		report.stat = "REJECTD"
	case 88:
		report.stat = "UNKNOWN"
	case 77:
		report.stat = "ACCEPTD"
	case 66:
		report.stat = "UNDELIV"
	case 55:
		report.stat = "DELETED"
	case 44:
		report.stat = "EXPIRED"
	case 33:
		report.stat = "MA:0000"
	case 22:
		report.stat = "MB:0000"
	case 11:
		report.stat = "CA:0000"
	case 10:
		report.stat = "CB:0000"
	default:
		report.stat = "DELIVRD"
	}
	return report
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
