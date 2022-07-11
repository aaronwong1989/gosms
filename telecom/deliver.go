package telecom

import (
	"encoding/binary"
	"fmt"
	"strings"
	"time"

	"sms-vgateway/comm"
)

type Deliver struct {
	*MessageHeader
	msgId      []byte        // 【10字节】短消息流水号
	isReport   byte          // 【1字节】是否为状态报告
	msgFormat  byte          // 【1字节】短消息格式
	recvTime   string        // 【14字节】短消息定时发送时间
	srcTermID  string        // 【21字节】短信息发送方号码
	destTermID string        // 【21】短消息接收号码
	msgLength  byte          // 【1字节】短消息长度
	msgContent string        // 【MsgLength字节】短消息内容
	msgBytes   []byte        // 消息内容按照Msg_Fmt编码后的数据
	report     *Report       // 状态报告
	reserve    string        // 【8字节】保留
	tlvList    *comm.TlvList // 【TLV】可选项参数
}

type DeliverResp struct {
	*MessageHeader
	msgId  []byte // 【10字节】短消息流水号
	status uint32
}

func NewDeliver(srcNo string, destNo string, txt string) *Deliver {
	baseLen := uint32(89)
	head := &MessageHeader{PacketLength: baseLen, RequestId: CmdDeliver, SequenceId: uint32(RequestSeq.NextVal())}
	dlv := &Deliver{MessageHeader: head}
	dlv.msgId = MsgIdSeq.NextSeq()
	dlv.isReport = 0
	dlv.msgFormat = 15
	dlv.recvTime = time.Now().Format("20060102150405")
	dlv.srcTermID = srcNo
	dlv.destTermID = Conf.DisplayNo + destNo
	// 上行最长70字符
	subTxt := txt
	rs := []rune(txt)
	if len(rs) > 70 {
		rs = rs[:70]
		subTxt = string(rs)
	}
	gbs, _ := GbEncoder.String(subTxt)
	msg := []byte(gbs)
	dlv.msgBytes = msg
	dlv.msgLength = byte(len(msg))
	dlv.msgContent = subTxt
	dlv.PacketLength = baseLen + uint32(dlv.msgLength)
	return dlv
}

func NewDeliveryReport(mt *Submit, msgId []byte) *Deliver {
	baseLen := uint32(89)
	head := &MessageHeader{PacketLength: baseLen, RequestId: CmdDeliver, SequenceId: uint32(RequestSeq.NextVal())}
	dlv := &Deliver{MessageHeader: head}
	rpt := NewReport(msgId)
	dlv.msgId = MsgIdSeq.NextSeq()
	dlv.report = rpt
	dlv.msgLength = 115
	dlv.isReport = 1
	dlv.msgFormat = 0
	dlv.recvTime = time.Now().Format("20060102150405")
	dlv.srcTermID = mt.destTermID[0]
	dlv.destTermID = mt.srcTermID
	dlv.PacketLength = baseLen + uint32(RptLen)
	return dlv
}

func (dlv *Deliver) Encode() []byte {
	frame := dlv.MessageHeader.Encode()
	index := 12
	copy(frame[index:index+10], dlv.msgId)
	index += 10
	index = comm.CopyByte(frame, dlv.isReport, index)
	index = comm.CopyByte(frame, dlv.msgFormat, index)
	index = comm.CopyStr(frame, dlv.recvTime, index, 14)
	index = comm.CopyStr(frame, dlv.srcTermID, index, 21)
	index = comm.CopyStr(frame, dlv.destTermID, index, 21)
	index = comm.CopyByte(frame, dlv.msgLength, index)
	if dlv.IsReport() && dlv.report != nil {
		rts := dlv.report.Encode()
		copy(frame[index:index+RptLen], rts)
		index += RptLen
	} else {
		copy(frame[index:index+int(dlv.msgLength)], dlv.msgBytes)
		index += int(dlv.msgLength)
	}
	index = comm.CopyStr(frame, dlv.reserve, index, 8)
	return frame
}

func (dlv *Deliver) Decode(header *MessageHeader, frame []byte) error {
	// check
	if header == nil || header.RequestId != CmdDeliver || uint32(len(frame)) < (header.PacketLength-HeadLength) {
		return ErrorPacket
	}
	dlv.MessageHeader = header
	var index int
	dlv.msgId = frame[index : index+10]
	index += 10
	dlv.isReport = frame[index]
	index += 1
	dlv.msgFormat = frame[index]
	index += 1
	dlv.recvTime = comm.TrimStr(frame[index : index+14])
	index += 14
	dlv.srcTermID = comm.TrimStr(frame[index : index+21])
	index += 21
	dlv.destTermID = comm.TrimStr(frame[index : index+21])
	index += 21
	dlv.msgLength = frame[index]
	index += 1
	if dlv.IsReport() {
		dlv.report = &Report{}
		err := dlv.report.Decode(frame[index : index+RptLen])
		if err != nil {
			return err
		}
	} else {
		bytes, err := GbDecoder.Bytes(frame[index : index+int(dlv.msgLength)])
		if err != nil {
			return err
		}
		dlv.msgContent = string(bytes)
	}
	// 后续字节不解析了
	return nil
}

func (dlv *Deliver) ToResponse(code uint32) interface{} {
	header := *dlv.MessageHeader
	header.RequestId = CmdDeliverResp
	header.PacketLength = 26
	resp := &DeliverResp{MessageHeader: &header}
	resp.status = code
	resp.msgId = MsgIdSeq.NextSeq()
	return resp
}

func (dlv *Deliver) String() string {
	content := ""
	if dlv.IsReport() {
		content = dlv.report.String()
	} else {
		content = strings.ReplaceAll(dlv.msgContent, "\n", " ")
	}
	return fmt.Sprintf("{ header: %v, msgId: %x, isReport: %v, msgFormat: %v, recvTime: %v,"+
		" srcTermID: %v, destTermID: %v, msgLength: %v, "+
		"msgContent: \"%s\", reserve: %v, tlv: %v }",
		dlv.MessageHeader, dlv.msgId, dlv.isReport, dlv.msgFormat, dlv.recvTime,
		dlv.srcTermID, dlv.destTermID, dlv.msgLength,
		content, dlv.reserve, dlv.tlvList,
	)
}

func (dlv *Deliver) IsReport() bool {
	return dlv.isReport == 1
}

func (r *DeliverResp) Encode() []byte {
	frame := r.MessageHeader.Encode()
	index := 12
	copy(frame[index:index+10], r.msgId)
	index += 10
	binary.BigEndian.PutUint32(frame[index:index+4], r.status)
	return frame
}

func (r *DeliverResp) Decode(header *MessageHeader, frame []byte) error {
	// check
	if header == nil || header.RequestId != CmdDeliverResp || uint32(len(frame)) < (header.PacketLength-HeadLength) {
		return ErrorPacket
	}
	r.MessageHeader = header
	r.msgId = make([]byte, 10)
	copy(r.msgId, frame[0:10])
	r.status = binary.BigEndian.Uint32(frame[10:14])
	return nil
}

func (r *DeliverResp) String() string {
	return fmt.Sprintf("{ header: %s, msgId: %x, status: {%d:%s} }", r.MessageHeader, r.msgId, r.status, StatMap[r.status])
}

func (r *DeliverResp) MsgId() string {
	return fmt.Sprintf("%x", r.msgId)
}
