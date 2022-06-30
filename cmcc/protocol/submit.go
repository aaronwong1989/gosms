package cmcc

import (
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

type SubmitResp struct {
	*MessageHeader
	msgId  uint64
	result uint32
}

func NewSubmit(phones []string, content string, opts ...Option) (messages []*Submit) {
	options := loadOptions(opts...)

	header := &MessageHeader{TotalLength: baseLen, CommandId: CMPP_SUBMIT, SequenceId: uint32(Sequence32.NextVal())}
	submit := &Submit{MessageHeader: header}

	setOptions(submit, options)
	setMsgFmt(content, submit)

	submit.destUsrTl = uint8(len(phones))
	submit.destTerminalId = strings.Join(phones, ",")
	termIds := make([]byte, 32*submit.destUsrTl)
	for i, p := range phones {
		copy(termIds[32*i:32*(i+1)], p)
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

func (s *Submit) ToResponse(result uint32) *SubmitResp {
	resp := &SubmitResp{}
	header := *s.MessageHeader
	resp.MessageHeader = &header
	resp.CommandId = CMPP_SUBMIT_RESP
	resp.TotalLength = HEAD_LENGTH + 12
	resp.msgId = uint64(Sequence64.NextVal())
	resp.result = result
	return resp
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
	return string(bts)
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
		s := time.Now().Add(Conf.ValidDuration).Format("20060102150304")
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

func (s *Submit) String() string {
	l := len(s.msgBytes)
	if l > 6 {
		l = 6
	}
	return fmt.Sprintf("{ header: %s, msgId: %v, pkTotal: %v, pkNumber: %v, registeredDel: %v, "+
		"msgLevel: %v, serviceId: %v, feeUsertype: %v, feeTerminalId: %v, feeTerminalType: %v, tpPid: %v, "+
		"tpUdhi: %v, msgFmt: %v, msgSrc: %v, feeType: %v, feeCode: %v, validTime: %v, atTime: %v, srcId: %v, "+
		"destUsrTl: %v, destTerminalId: [%v], destTerminalType: %v, msgLength: %v, msgBytes: %0X..., linkID: %v }",
		s.MessageHeader, s.msgId, s.pkTotal, s.pkNumber, s.registeredDel,
		s.msgLevel, s.serviceId, s.feeUsertype, s.feeTerminalId, s.feeTerminalType, s.tpPid,
		s.tpUdhi, s.msgFmt, s.msgSrc, s.feeType, s.feeCode, s.validTime, s.atTime, s.srcId,
		s.destUsrTl, s.destTerminalId, s.destTerminalType, s.msgLength, s.msgBytes[0:l], s.linkID,
	)
}

func (s *SubmitResp) String() string {
	return fmt.Sprintf("{ header: %s, msgId: %v, result: %v }", s.MessageHeader, s.msgId, s.result)
}
