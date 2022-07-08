package telecom

import (
	"encoding/binary"
	"fmt"
	"time"

	"sms-vgateway/tool"
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
	tlvList    *tool.TlvList // 【TLV】可选项参数
}

type DeliverResp struct {
	*MessageHeader
	msgId  []byte // 【10字节】短消息流水号
	status uint32
}

func NewDeliver(srcNo string, destNo string, txt string) *Deliver {
	baseLen := uint32(89)
	head := &MessageHeader{PacketLength: baseLen, RequestId: CmdDeliver, SequenceId: uint32(Sequence32.NextVal())}
	dlv := &Deliver{MessageHeader: head}
	dlv.isReport = 0
	dlv.msgFormat = 15
	dlv.recvTime = time.Now().Format("20060102150405")
	dlv.srcTermID = srcNo
	dlv.destTermID = destNo
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
	return dlv
}

func (dlv *Deliver) Deliver() []byte {
	return nil
}

func (dlv *Deliver) Decode(header *MessageHeader, frame []byte) error {
	return nil
}

func (dlv *Deliver) ToResponse(code uint32) interface{} {
	header := *dlv.MessageHeader
	header.RequestId = CmdSubmitResp
	header.PacketLength = 26
	resp := &SubmitResp{MessageHeader: &header}
	resp.status = code
	resp.msgId = MsgIdSeq.NextSeq()
	return resp
}

func (dlv *Deliver) String() string {
	return ""
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
	if header == nil || header.RequestId != CmdSubmitResp || uint32(len(frame)) < (header.PacketLength-HeadLength) {
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
