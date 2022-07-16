package cmpp

import (
	"encoding/binary"
	"fmt"
	"strings"
	"time"

	"gosms/comm"
)

// Submit
// 3.0版 feeTerminalId、destTerminalId 均为32字节，无Reserve字段，有LinkId字段
// 2.0版 feeTerminalId、destTerminalId 均为21字节，无LinkId字段，有Reserve字段

type Submit struct {
	*MessageHeader        // 消息头，【12字节】
	msgId          uint64 // 信息标识，由 SP 接入的短信网关本身产 生，本处填空(0)。【8字节】
	pkTotal        uint8  // 相同Msg_Id的信息总条数 【1字节】
	pkNumber       uint8  // 相同Msg_Id的信息序号，从1开始 【1字节】
	registeredDel  uint8  // 是否要求返回状态确认报告： 0：不需要，1：需要。【1字节】
	msgLevel       uint8  // 信息级别，1-9 【1字节】
	serviceId      string // 业务标识，是数字、字母和符号的组合。【10字节】
	// 计费用户类型字段
	// 0：对目的终端MSISDN计费；
	// 1：对源终端MSISDN计费；
	// 2：对SP计费;
	// 3：表示本字段无效，对谁计费参见Fee_terminal_Id 字段。
	feeUsertype uint8 // 【1字节】
	// 被计费用户的号码（如本字节填空，则表示本字段无效，对谁计费参见Fee_UserType字段，本字段与Fee_UserType字段互斥）
	feeTerminalId   string //  【32字节】
	feeTerminalType uint8  // 被计费用户的号码类型，0：真实号码；1：伪码 【1字节】
	tpPid           uint8  // GSM协议类型。详细是解释请参考GSM03.40中的9.2.3.9 【1字节】
	tpUdhi          uint8  // GSM协议类型。详细是解释请参考GSM03.40中的9.2.3.9 【1字节】
	// 信息格式
	// 0：ASCII串
	// 3：短信写卡操作
	// 4：二进制信息
	// 8：UCS2编码
	// 15：含GB汉字
	msgFmt uint8  //  【1字节】
	msgSrc string // 信息内容来源(SP_Id) 【6字节】
	// 资费类别
	// 01：对“计费用户号码”免费
	// 02：对“计费用户号码”按条计信息费
	// 03：对“计费用户号码”按包月收取信息费
	// 04：对“计费用户号码”的信息费封顶
	// 05：对“计费用户号码”的收费是由SP实现
	feeType   string //  【2字节】
	feeCode   string // 资费代码（以分为单位） 【6字节】
	validTime string // 存活有效期，格式遵循SMPP3.3协议 【17字节】
	atTime    string // 定时发送时间，格式遵循SMPP3.3协议 【17字节】
	// 源号码 SP的服务代码或前缀为服务代码的长号码, 网关将该号码完整的填到SMPP协议Submit_SM消息相应的source_addr字段，该号码最终在用户手机上显示为短消息的主叫号码
	srcId     string //  【21字节】
	destUsrTl uint8  // 接收信息的用户数量(小于100个用户) 【1字节】
	// 接收短信的MSISDN号码
	destTerminalId   string //  【32*DestUsrTl字节】
	termIds          []byte // DestTerminalId编码后的格式
	destTerminalType uint8  //  接收短信的用户的号码类型，0：真实号码；1：伪码【1字节】
	msgLength        uint8  // 信息长度(Msg_Fmt值为0时：<160个字节；其它<=140个字节) 【1字节】
	msgContent       string // 信息内容 【MsgLength字节】
	msgBytes         []byte // 消息内容按照Msg_Fmt编码后的数据
	linkID           string // 点播业务使用的LinkID，非点播类业务的MT流程不使用该字段 【20字节】
}

func NewSubmit(phones []string, content string, opts ...Option) (messages []*Submit) {
	options := loadOptions(opts...)
	baseLen := 138
	if V3() {
		baseLen = 163
	}
	header := &MessageHeader{TotalLength: uint32(baseLen), CommandId: CMPP_SUBMIT, SequenceId: uint32(Seq32.NextVal())}
	mt := &Submit{MessageHeader: header}

	setOptions(mt, options)
	mt.msgFmt = MsgFmt(content)

	mt.destUsrTl = uint8(len(phones))
	mt.destTerminalId = strings.Join(phones, ",")
	idLen := 21
	if V3() {
		idLen = 32
	}
	termIds := make([]byte, idLen*int(mt.destUsrTl))
	for i, p := range phones {
		copy(termIds[i*idLen:(i+1)*idLen], p)
	}
	mt.termIds = termIds

	mt.msgSrc = Conf.SourceAddr

	mt.msgContent = content
	slices := MsgSlices(mt.msgFmt, content)

	if len(slices) == 1 {
		mt.pkTotal = 1
		mt.pkNumber = 1
		mt.msgLength = uint8(len(slices[0]))
		mt.msgBytes = slices[0]
		mt.TotalLength = uint32(baseLen + len(termIds) + len(slices[0]))
		return []*Submit{mt}
	} else {
		mt.tpUdhi = 1
		mt.pkTotal = uint8(len(slices))

		for i, msgBytes := range slices {
			// 拷贝 mt
			tmp := *mt
			tmpHead := *tmp.MessageHeader
			sub := &tmp
			sub.MessageHeader = &tmpHead
			if i != 0 {
				sub.SequenceId = uint32(Seq32.NextVal())
			}
			sub.pkNumber = uint8(i + 1)
			sub.msgLength = uint8(len(msgBytes))
			sub.msgBytes = msgBytes
			sub.TotalLength = uint32(baseLen + len(termIds) + len(msgBytes))
			messages = append(messages, sub)
		}

		return messages
	}
}

func (sub *Submit) Encode() []byte {
	frame := sub.MessageHeader.Encode()
	frame[20] = sub.pkTotal
	frame[21] = sub.pkNumber
	frame[22] = sub.registeredDel
	frame[23] = sub.msgLevel
	copy(frame[24:34], sub.serviceId)
	frame[34] = sub.feeUsertype
	index := 35
	if V3() {
		copy(frame[index:index+32], sub.feeTerminalId)
		index += 32
		frame[index] = sub.feeTerminalType
		index++
	} else {
		copy(frame[index:index+21], sub.feeTerminalId)
		index += 21
	}
	frame[index] = sub.tpPid
	index++
	frame[index] = sub.tpUdhi
	index++
	frame[index] = sub.msgFmt
	index++
	copy(frame[index:index+6], sub.msgSrc)
	index += 6
	copy(frame[index:index+2], sub.feeType)
	index += 2
	copy(frame[index:index+6], sub.feeCode)
	index += 6
	copy(frame[index:index+17], sub.validTime)
	index += 17
	copy(frame[index:index+17], sub.atTime)
	index += 17
	copy(frame[index:index+21], sub.srcId)
	index += 21
	frame[index] = sub.destUsrTl
	index++
	copy(frame[index:index+len(sub.termIds)], sub.termIds)
	index += len(sub.termIds)
	if V3() {
		frame[index] = sub.destTerminalType
		index++
	}
	frame[index] = sub.msgLength
	index++
	copy(frame[index:index+len(sub.msgBytes)], sub.msgBytes)
	index += len(sub.msgBytes)
	if V3() {
		copy(frame[index:index+20], sub.linkID)
	}
	return frame
}

func (sub *Submit) Decode(header *MessageHeader, frame []byte) error {
	// check
	if header == nil || header.CommandId != CMPP_SUBMIT || uint32(len(frame)) < (header.TotalLength-HEAD_LENGTH) {
		return ErrorPacket
	}
	sub.MessageHeader = header
	// msgId uint64
	index := 8
	sub.pkTotal = frame[index]
	index++
	sub.pkNumber = frame[index]
	index++
	sub.registeredDel = frame[index]
	index++
	sub.msgLevel = frame[index]
	index++
	sub.serviceId = TrimStr(frame[index : index+10])
	index += 10
	sub.feeUsertype = frame[index]
	index++
	if V3() {
		sub.feeTerminalId = TrimStr(frame[index : index+32])
		index += 32
		sub.feeTerminalType = frame[index]
		index++
	} else {
		sub.feeTerminalId = TrimStr(frame[index : index+21])
		index += 21
	}
	sub.tpPid = frame[index]
	index++
	sub.tpUdhi = frame[index]
	index++
	sub.msgFmt = frame[index]
	index++
	sub.msgSrc = TrimStr(frame[index : index+6])
	index += 6
	sub.feeType = TrimStr(frame[index : index+2])
	index += 2
	sub.feeCode = TrimStr(frame[index : index+6])
	index += 6
	sub.validTime = TrimStr(frame[index : index+17])
	index += 17
	sub.atTime = TrimStr(frame[index : index+17])
	index += 17
	sub.srcId = TrimStr(frame[index : index+21])
	index += 21
	sub.destUsrTl = frame[index]
	index++
	l := int(sub.destUsrTl * 21)
	if V3() {
		l = int(sub.destUsrTl) << 5
	}
	sub.destTerminalId = TrimStr(frame[index : index+l])
	index += l
	if V3() {
		sub.destTerminalType = frame[index]
		index++
	}
	sub.msgLength = frame[index]
	index++
	content := frame[index : index+int(sub.msgLength)]
	sub.msgBytes = content
	if content[0] == 0x05 && content[1] == 0x00 && content[2] == 0x03 {
		content = content[6:]
	}
	if sub.msgFmt == 8 {
		sub.msgContent = comm.Ucs2Decode(content)
	} else {
		sub.msgContent = TrimStr(content)
	}
	index += int(sub.msgLength)
	if V3() {
		sub.linkID = TrimStr(frame[index : index+20])
	}
	return nil
}

type SubmitResp struct {
	*MessageHeader
	msgId  uint64
	result uint32
}

func (sub *Submit) ToResponse(result uint32) interface{} {
	resp := &SubmitResp{}
	header := *sub.MessageHeader
	resp.MessageHeader = &header
	resp.CommandId = CMPP_SUBMIT_RESP
	resp.TotalLength = HEAD_LENGTH + 9
	if V3() {
		resp.TotalLength = HEAD_LENGTH + 12
	}
	if result == 0 {
		resp.msgId = uint64(Seq64.NextVal())
	}
	resp.result = result
	return resp
}

func (sub *Submit) ToDeliveryReport(msgId uint64) *Delivery {
	d := Delivery{}

	head := *sub.MessageHeader
	d.MessageHeader = &head
	d.TotalLength = 145
	if V3() {
		d.TotalLength = 169
	}
	d.CommandId = CMPP_DELIVER
	d.SequenceId = uint32(Seq32.NextVal())

	d.registeredDelivery = 1
	d.msgLength = 60
	d.destId = sub.srcId
	d.serviceId = sub.serviceId
	d.srcTerminalId = sub.destTerminalId
	d.srcTerminalType = sub.destTerminalType

	subTime := time.Now().Format("0601021504")
	doneTime := time.Now().Add(10 * time.Second).Format("0601021504")
	report := NewReport(msgId, sub.destTerminalId, subTime, doneTime)
	d.report = report

	return &d
}

func (resp *SubmitResp) Encode() []byte {
	frame := resp.MessageHeader.Encode()
	binary.BigEndian.PutUint64(frame[12:20], resp.msgId)
	if V3() {
		binary.BigEndian.PutUint32(frame[20:24], resp.result)
	} else {
		frame[20] = byte(resp.result)
	}
	return frame
}
func (resp *SubmitResp) Decode(header *MessageHeader, frame []byte) error {
	// check
	if header == nil || header.CommandId != CMPP_SUBMIT_RESP || uint32(len(frame)) < (header.TotalLength-HEAD_LENGTH) {
		return ErrorPacket
	}
	resp.MessageHeader = header
	resp.msgId = binary.BigEndian.Uint64(frame[0:8])
	if V3() {
		resp.result = binary.BigEndian.Uint32(frame[8:12])
	} else {
		resp.result = uint32(frame[8])
	}
	return nil
}

func (resp *SubmitResp) MsgId() uint64 {
	return resp.msgId
}

func (resp *SubmitResp) Result() uint32 {
	return resp.result
}

func MsgSlices(fmt uint8, content string) (slices [][]byte) {
	var msgBytes []byte
	// 含中文
	if fmt == 8 {
		msgBytes = comm.Ucs2Encode(content)
		slices = comm.ToTPUDHISlices(msgBytes, 140)
	} else {
		// 纯英文
		msgBytes = []byte(content)
		slices = comm.ToTPUDHISlices(msgBytes, 160)
	}
	return
}

// MsgFmt 通过消息内容判断，设置编码格式。
// 如果是纯拉丁字符采用0：ASCII串
// 如果含多字节字符，这采用8：UCS-2编码
func MsgFmt(content string) uint8 {
	if len(content) < 2 {
		return 0
	}
	all7bits := len(content) == len([]rune(content))
	if all7bits {
		return 0
	} else {
		return 8
	}
}

// 设置可选项
func setOptions(sub *Submit, opts *MtOptions) {
	if opts.FeeUsertype != uint8(0xf) {
		sub.feeUsertype = opts.FeeUsertype
	} else {
		sub.feeUsertype = Conf.FeeUsertype
	}

	if opts.MsgLevel != uint8(0xf) {
		sub.msgLevel = opts.MsgLevel
	} else {
		sub.msgLevel = Conf.MsgLevel
	}

	if opts.RegisteredDel != uint8(0xf) {
		sub.registeredDel = opts.RegisteredDel
	} else {
		sub.registeredDel = Conf.RegisteredDel
	}

	if opts.FeeTerminalType != uint8(0xf) {
		sub.feeTerminalType = opts.FeeTerminalType
	} else {
		sub.feeTerminalType = Conf.FeeTerminalType
	}

	if opts.FeeType != "" {
		sub.feeType = opts.FeeType
	} else {
		sub.feeType = Conf.FeeType
	}

	if opts.AtTime != "" {
		sub.atTime = opts.AtTime
	}

	if opts.ValidTime != "" {
		sub.validTime = opts.ValidTime
	} else {
		t := time.Now().Add(Conf.ValidDuration)
		s := t.Format("060102150405")
		sub.validTime = s + "032+"
	}

	if opts.FeeCode != "" {
		sub.feeCode = opts.FeeCode
	} else {
		sub.feeCode = Conf.FeeCode
	}

	if opts.FeeTerminalId != "" {
		sub.feeTerminalId = opts.FeeTerminalId
	} else {
		sub.feeTerminalId = Conf.FeeTerminalId
	}

	if opts.SrcId != "" {
		sub.srcId = opts.SrcId
	} else {
		sub.srcId = Conf.SrcId
	}

	if opts.ServiceId != "" {
		sub.serviceId = opts.ServiceId
	} else {
		sub.serviceId = Conf.ServiceId
	}

	if opts.LinkID != "" {
		sub.linkID = opts.LinkID
	} else {
		sub.linkID = Conf.LinkID
	}
}

func (sub *Submit) String() string {
	l := len(sub.msgBytes)
	if l > 6 {
		l = 6
	}
	return fmt.Sprintf("{ header: %s, msgId: %v, pkTotal: %v, pkNumber: %v, registeredDel: %v, "+
		"msgLevel: %v, serviceId: %v, feeUsertype: %v, feeTerminalId: %v, feeTerminalType: %v, tpPid: %v, "+
		"tpUdhi: %v, msgFmt: %v, msgSrc: %v, feeType: %v, feeCode: %v, validTime: %v, atTime: %v, srcId: %v, "+
		"destUsrTl: %v, destTerminalId: [%v], destTerminalType: %v, msgLength: %v, msgBytes: %0x..., linkID: %v }",
		sub.MessageHeader, sub.msgId, sub.pkTotal, sub.pkNumber, sub.registeredDel,
		sub.msgLevel, sub.serviceId, sub.feeUsertype, sub.feeTerminalId, sub.feeTerminalType, sub.tpPid,
		sub.tpUdhi, sub.msgFmt, sub.msgSrc, sub.feeType, sub.feeCode, sub.validTime, sub.atTime, sub.srcId,
		sub.destUsrTl, sub.destTerminalId, sub.destTerminalType, sub.msgLength, sub.msgBytes[0:l], sub.linkID,
	)
}

func (resp *SubmitResp) String() string {
	return fmt.Sprintf("{ header: %s, msgId: %v, result: {%d: %s} }", resp.MessageHeader, resp.msgId, resp.result, SubmitResultMap[resp.result])
}

var SubmitResultMap = map[uint32]string{
	0:  "正确",
	1:  "消息结构错",
	2:  "命令字错",
	3:  "消息序号重复",
	4:  "消息长度错",
	5:  "资费代码错",
	6:  "超过最大信息长",
	7:  "业务代码错",
	8:  "流量控制错",
	9:  "本网关不负责服务此计费号码",
	10: "Src_Id 错误",
	11: "Msg_src 错误",
	12: "Fee_terminal_Id 错误",
	13: "Dest_terminal_Id 错误",
}
