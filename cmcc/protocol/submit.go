package cmcc

import (
	"encoding/binary"
	"fmt"
	"strings"
	"time"

	"golang.org/x/text/encoding/unicode"
	"golang.org/x/text/transform"
)

const (
	baseLen = 163
)

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

	header := &MessageHeader{TotalLength: baseLen, CommandId: CMPP_SUBMIT, SequenceId: uint32(Sequence32.NextVal())}
	submit := &Submit{MessageHeader: header}

	setOptions(submit, options)
	setMsgFmt(content, submit)

	submit.destUsrTl = uint8(len(phones))
	submit.destTerminalId = strings.Join(phones, ",")
	termIds := make([]byte, int(submit.destUsrTl)<<5)
	for i, p := range phones {
		copy(termIds[i<<5:(i+1)<<5], p)
	}
	submit.termIds = termIds

	submit.msgSrc = Conf.SourceAddr

	submit.msgContent = content
	slices := msgSlices(submit.msgFmt, content)

	if len(slices) == 1 {
		submit.pkTotal = 1
		submit.pkNumber = 1
		submit.msgLength = uint8(len(slices[0]))
		submit.msgBytes = slices[0]
		submit.TotalLength = uint32(baseLen + len(termIds) + len(slices[0]))
		return []*Submit{submit}
	} else {
		submit.tpUdhi = 1
		submit.pkTotal = uint8(len(slices))

		for i, msgBytes := range slices {
			// 拷贝 submit
			tmp := *submit
			tmpHead := *tmp.MessageHeader
			sub := &tmp
			sub.MessageHeader = &tmpHead

			sub.SequenceId = uint32(Sequence32.NextVal())
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
	// msgId +8 = 20
	// Pk_total + 1 = 21
	frame[20] = sub.pkTotal
	// Pk_number + 1 = 22
	frame[21] = sub.pkNumber
	// Registered_Delivery + 1 = 23
	frame[22] = sub.registeredDel
	// Msg_level + 1 = 24
	frame[23] = sub.msgLevel
	// Service_Id + 10 = 34
	copy(frame[24:34], sub.serviceId)
	// Fee_UserType + 1 = 35
	frame[34] = sub.feeUsertype
	// Fee_terminal_Id + 32 = 67
	copy(frame[35:67], sub.feeTerminalId)
	// Fee_terminal_type + 1 = 68
	frame[67] = sub.feeTerminalType
	// TP_pId	    + 1  = 69
	frame[68] = sub.tpPid
	// TP_udhi	  + 1  = 70
	frame[69] = sub.tpUdhi
	// Msg_Fmt	  + 1  = 71
	frame[70] = sub.msgFmt
	// Msg_src	  + 6  = 77
	copy(frame[71:77], sub.msgSrc)
	// FeeType	  + 2  = 79
	copy(frame[77:79], sub.feeType)
	// FeeCode	  + 6  = 85
	copy(frame[79:85], sub.feeCode)
	// ValId_Time	   + 17  = 102
	copy(frame[85:102], sub.validTime)
	// At_Time	     + 17  = 119
	copy(frame[102:119], sub.atTime)
	// Src_Id	       + 21  = 140
	copy(frame[119:140], sub.srcId)
	// DestUsr_tl	   + 1   = 141
	frame[140] = sub.destUsrTl
	// Dest_terminal_Id	     + 32*DestUsr_tl
	index := 141 + len(sub.termIds)
	copy(frame[141:index], sub.termIds)
	// Dest_terminal_type	   + 1
	frame[index] = sub.destTerminalType
	index++
	// Msg_Length	           + 1
	frame[index] = sub.msgLength
	index++
	// Msg_Content	         + Msg_length
	copy(frame[index:index+len(sub.msgBytes)], sub.msgBytes)
	index += len(sub.msgBytes)
	// LinkID 	             + 20
	copy(frame[index:index+20], sub.linkID)
	return frame
}

func (sub *Submit) Decode(header *MessageHeader, frame []byte) error {
	// check
	if header == nil || header.CommandId != CMPP_SUBMIT || uint32(len(frame)) < (header.TotalLength-HEAD_LENGTH) {
		return ErrorPacket
	}
	sub.MessageHeader = header
	// msgId +8 = 8
	// Pk_total + 1 = 9
	sub.pkTotal = frame[8]
	// Pk_number + 1 = 10
	sub.pkNumber = frame[9]
	// Registered_Delivery + 1 = 11
	sub.registeredDel = frame[10]
	// Msg_level + 1 = 12
	sub.msgLevel = frame[11]
	// Service_Id + 10 = 22
	sub.serviceId = TrimStr(frame[12:22])
	//  Fee_UserType + 1 = 23
	sub.feeUsertype = frame[22]
	// Fee_terminal_Id + 32 = 55
	sub.feeTerminalId = TrimStr(frame[23:55])
	// Fee_terminal_type + 1 = 56
	sub.feeTerminalType = frame[55]
	// TP_pId	    + 1  = 57
	sub.tpPid = frame[56]
	// TP_udhi	  + 1  = 58
	sub.tpUdhi = frame[57]
	// Msg_Fmt	  + 1  = 59
	sub.msgFmt = frame[58]
	// Msg_src	  + 6  = 65
	sub.msgSrc = TrimStr(frame[59:65])
	// FeeType	  + 2  = 67
	sub.feeType = TrimStr(frame[65:67])
	// FeeCode	  + 6  = 73
	sub.feeCode = TrimStr(frame[67:73])
	// ValId_Time	   + 17  = 90
	sub.validTime = TrimStr(frame[73:90])
	// At_Time	     + 17  = 107
	sub.atTime = TrimStr(frame[90:107])
	// Src_Id	       + 21  = 128
	sub.srcId = TrimStr(frame[107:128])
	// DestUsr_tl	   + 1   = 129
	sub.destUsrTl = frame[128]
	// Dest_terminal_Id	     + 32*DestUsr_tl
	index := 129 + int(sub.destUsrTl)<<5
	sub.destTerminalId = TrimStr(frame[129:index])
	// Dest_terminal_type	   + 1
	sub.destTerminalType = frame[index]
	index++
	// Msg_Length	           + 1
	sub.msgLength = frame[index]
	index++
	// Msg_Content	         + Msg_length
	content := frame[index : index+int(sub.msgLength)]
	sub.msgBytes = content
	if content[0] == 0x05 && content[1] == 0x00 && content[2] == 0x03 {
		content = content[6:]
	}
	if sub.msgFmt == 8 {
		sub.msgContent = ucs2Decode(content)
	} else {
		sub.msgContent = TrimStr(content)
	}
	index += int(sub.msgLength)
	// LinkID 	             + 20
	sub.linkID = TrimStr(frame[index : index+20])
	return nil
}

type SubmitResp struct {
	*MessageHeader
	msgId  uint64
	result uint32
}

func (sub *Submit) ToResponse(result uint32) *SubmitResp {
	resp := &SubmitResp{}
	header := *sub.MessageHeader
	resp.MessageHeader = &header
	resp.CommandId = CMPP_SUBMIT_RESP
	resp.TotalLength = HEAD_LENGTH + 12
	resp.msgId = uint64(Sequence64.NextVal())
	resp.result = result
	return resp
}

func (resp *SubmitResp) Encode() []byte {
	frame := resp.MessageHeader.Encode()
	binary.BigEndian.PutUint64(frame[12:20], resp.msgId)
	binary.BigEndian.PutUint32(frame[20:24], resp.result)
	return frame
}
func (resp *SubmitResp) Decode(header *MessageHeader, frame []byte) error {
	// check
	if header == nil || header.CommandId != CMPP_SUBMIT_RESP || uint32(len(frame)) < (header.TotalLength-HEAD_LENGTH) {
		return ErrorPacket
	}
	resp.MessageHeader = header
	resp.msgId = binary.BigEndian.Uint64(frame[0:8])
	resp.result = binary.BigEndian.Uint32(frame[8:12])
	return nil
}

func msgSlices(fmt uint8, content string) (slices [][]byte) {
	var msgBytes []byte
	// 含中文
	if fmt == 8 {
		msgBytes = ucs2Encode(content)
		if len(msgBytes) <= 140 {
			slices = [][]byte{msgBytes}
		} else {
			slices = toTpUdhiSlices(msgBytes, 140)
		}
	} else {
		// 纯英文
		msgBytes = []byte(content)
		if len(msgBytes) <= 160 {
			slices = [][]byte{msgBytes}
		} else {
			slices = toTpUdhiSlices(msgBytes, 160)
		}
	}
	return
}

// 拆分为长短信切片
// 纯ASCII内容的拆分 pkgLen = 160
// 含中文内容的拆分   pkgLen = 140
func toTpUdhiSlices(content []byte, pkgLen int) (rt [][]byte) {
	headLen := 6
	bodyLen := pkgLen - headLen
	parts := len(content) / bodyLen
	tailLen := len(content) % bodyLen
	if tailLen != 0 {
		parts++
	}
	// 分片消息组的标识，用于收集组装消息
	groupId := byte(time.Now().UnixNano() & 0xff)
	var part []byte
	for i := 0; i < parts; i++ {
		if i != parts-1 {
			part = make([]byte, pkgLen)
		} else {
			// 最后一片
			part = make([]byte, 6+tailLen)
		}
		part[0], part[1], part[2] = 0x05, 0x00, 0x03
		part[3] = groupId
		part[4], part[5] = byte(parts), byte(i+1)
		if i != parts-1 {
			copy(part[6:pkgLen], content[bodyLen*i:bodyLen*(i+1)])
		} else {
			copy(part[6:], content[0:tailLen])
		}
		rt = append(rt, part)
	}
	return rt
}

// Encode to UCS2.
func ucs2Encode(s string) []byte {
	e := unicode.UTF16(unicode.BigEndian, unicode.IgnoreBOM)
	ucs, _, err := transform.Bytes(e.NewEncoder(), []byte(s))
	if err != nil {
		return nil
	}
	return ucs
}

// Decode from UCS2.
func ucs2Decode(ucs2 []byte) string {
	e := unicode.UTF16(unicode.BigEndian, unicode.IgnoreBOM)
	bts, _, err := transform.Bytes(e.NewDecoder(), ucs2)
	if err != nil {
		return ""
	}
	return TrimStr(bts)
}

// 通过消息内容判断，设置编码格式。
// 如果是纯拉丁字符采用0：ASCII串
// 如果含多字节字符，这采用8：UCS-2编码
func setMsgFmt(content string, submit *Submit) {
	all7bits := len(content) == len([]rune(content))
	if all7bits {
		submit.msgFmt = 0
	} else {
		submit.msgFmt = 8
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
