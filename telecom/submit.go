package telecom

import (
	"bytes"

	"sms-vgateway/tool"
)

type Submit struct {
	*MessageHeader
	msgType         byte          // 【1字节】短消息类型
	needReport      byte          // 【1字节】SP是否要求返回状态报告
	priority        byte          // 【1字节】短消息发送优先级
	serviceID       string        // 【10字节】业务代码
	feeType         string        // 【2字节】收费类型
	feeCode         string        // 【6字节】资费代码
	fixedFee        string        // 【6字节】包月费/封顶费
	msgFormat       byte          // 【1字节】短消息格式
	atTime          string        // 【17字节】短消息定时发送时间
	validTime       string        // 【17字节】短消息有效时间
	srcTermID       string        // 【21字节】短信息发送方号码
	chargeTermID    string        // 【21字节】计费用户号码
	destTermIDCount byte          // 【1字节】短消息接收号码总数
	destTermID      []string      // 【21*DestTermCount字节】短消息接收号码
	msgLength       byte          // 【1字节】短消息长度
	msgContent      string        // 【MsgLength字节】短消息内容
	msgBytes        []byte        // 消息内容按照Msg_Fmt编码后的数据
	reserve         string        // 【8字节】保留
	tlvList         *tool.TlvList // 【TLV】可选项参数
}

const MtBaseLen = 126

func NewSubmit(phones []string, content string, options MtOptions) (messages []*Submit) {

	head := &MessageHeader{PacketLength: MtBaseLen, RequestId: CmdSubmit, SequenceId: uint32(Sequence32.NextVal())}
	mt := &Submit{}
	mt.MessageHeader = head
	mt.SetOptions(options)

	// 从配置文件设置属性
	mt.feeType = Conf.FeeType
	mt.feeCode = Conf.FeeCode
	mt.chargeTermID = Conf.ChargeTermID
	mt.fixedFee = Conf.FixedFee
	// 初步设置入参
	mt.destTermID = phones
	mt.destTermIDCount = byte(len(phones))

	mt.msgFormat = 15
	data, err := GbEncoder.Bytes([]byte(content))
	if err != nil {
		return nil
	}
	slices := tool.ToTPUDHISlices(data, 140)
	if len(slices) == 1 {
		mt.msgBytes = slices[0]
		mt.msgLength = byte(len(mt.msgBytes))
		mt.PacketLength = uint32(MtBaseLen + len(mt.destTermID)*21 + int(mt.msgLength))
		return []*Submit{mt}
	} else {
		for i, dt := range slices {
			// 拷贝 mt
			tmp := *mt
			tmpHead := *tmp.MessageHeader
			sub := &tmp
			sub.MessageHeader = &tmpHead
			if i != 0 {
				sub.SequenceId = uint32(Sequence32.NextVal())
			}
			sub.msgLength = byte(len(dt))
			sub.msgBytes = dt
			l := 0
			sub.tlvList = tool.NewTlvList()
			sub.tlvList.Add(TP_pid, []byte{0x01})
			l += 5
			sub.tlvList.Add(TP_udhi, []byte{0x01})
			l += 5
			sub.tlvList.Add(PkTotal, []byte{byte(len(slices))})
			l += 5
			sub.tlvList.Add(PkNumber, []byte{byte(i)})
			l += 5
			sub.PacketLength = uint32(MtBaseLen + len(sub.destTermID)*21 + int(sub.msgLength) + l)
			messages = append(messages, sub)
		}
		return messages
	}
}

func (s *Submit) Encode() []byte {
	if len(s.destTermID) != int(s.destTermIDCount) {
		return nil
	}
	frame := s.MessageHeader.Encode()
	index := 12
	index = tool.CopyByte(frame, s.msgType, index)
	index = tool.CopyByte(frame, s.needReport, index)
	index = tool.CopyByte(frame, s.priority, index)
	index = tool.CopyStr(frame, s.serviceID, index, 10)
	index = tool.CopyStr(frame, s.feeType, index, 2)
	index = tool.CopyStr(frame, s.feeCode, index, 6)
	index = tool.CopyStr(frame, s.fixedFee, index, 6)
	index = tool.CopyByte(frame, s.msgFormat, index)
	index = tool.CopyStr(frame, s.atTime, index, 17)
	index = tool.CopyStr(frame, s.validTime, index, 17)
	index = tool.CopyStr(frame, s.srcTermID, index, 21)
	index = tool.CopyStr(frame, s.chargeTermID, index, 21)
	index = tool.CopyByte(frame, s.destTermIDCount, index)
	for _, tid := range s.destTermID {
		index = tool.CopyStr(frame, tid, index, 21)
	}

	index = tool.CopyByte(frame, s.msgLength, index)
	copy(frame[index:index+int(s.msgLength)], s.msgBytes)
	index += +int(s.msgLength)
	index = tool.CopyStr(frame, s.reserve, index, 8)
	buff := new(bytes.Buffer)
	err := s.tlvList.Write(buff)
	if err != nil {
		log.Errorf("%v", err)
		return nil
	}
	copy(frame[index:], buff.Bytes())
	return frame
}

func (s *Submit) Decode(header *MessageHeader, frame []byte) error {
	// TODO implement me
	panic("implement me")
}

func (s *Submit) ToResponse(code uint32) interface{} {
	// TODO implement me
	panic("implement me")
}

const (
	TP_pid           = uint16(0x0001)
	TP_udhi          = uint16(0x0002)
	LinkID           = uint16(0x0003)
	ChargeUserType   = uint16(0x0004)
	ChargeTermType   = uint16(0x0005)
	ChargeTermPseudo = uint16(0x0006)
	DestTermType     = uint16(0x0007)
	DestTermPseudo   = uint16(0x0008)
	PkTotal          = uint16(0x0009)
	PkNumber         = uint16(0x000A)
	SubmitMsgType    = uint16(0x000B)
	SPDealReslt      = uint16(0x000C)
	SrcTermType      = uint16(0x000D)
	SrcTermPseudo    = uint16(0x000E)
	NodesCount       = uint16(0x000F)
	MsgSrc           = uint16(0x0010)
	SrcType          = uint16(0x0011)
	MServiceID       = uint16(0x0012)
)