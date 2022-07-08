package telecom

import (
	"bytes"
	"encoding/binary"
	"fmt"

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

type SubmitResp struct {
	*MessageHeader
	msgId  []byte // 【10字节】短消息流水号
	status uint32
}

const MtBaseLen = 126

func NewSubmit(phones []string, content string, options MtOptions) (messages []*Submit) {

	head := &MessageHeader{PacketLength: MtBaseLen, RequestId: CmdSubmit, SequenceId: uint32(Sequence32.NextVal())}
	mt := &Submit{}
	mt.MessageHeader = head
	mt.SetOptions(options)
	mt.msgType = 6
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
	if s.tlvList != nil {
		buff := new(bytes.Buffer)
		err := s.tlvList.Write(buff)
		if err != nil {
			log.Errorf("%v", err)
			return nil
		}
		copy(frame[index:], buff.Bytes())
	}
	return frame
}

func (s *Submit) Decode(header *MessageHeader, frame []byte) error {
	// check
	if header == nil || header.RequestId != CmdSubmit || uint32(len(frame)) < (header.PacketLength-HeadLength) {
		return ErrorPacket
	}
	s.MessageHeader = header

	var index int
	s.msgType = frame[index]
	index++
	s.needReport = frame[index]
	index++
	s.priority = frame[index]
	index++
	s.serviceID = tool.TrimStr(frame[index : index+10])
	index += 10
	s.feeType = tool.TrimStr(frame[index : index+2])
	index += 2
	s.feeCode = tool.TrimStr(frame[index : index+6])
	index += 6
	s.fixedFee = tool.TrimStr(frame[index : index+6])
	index += 6
	s.msgFormat = frame[index]
	index++
	s.atTime = tool.TrimStr(frame[index : index+17])
	index += 17
	s.validTime = tool.TrimStr(frame[index : index+17])
	index += 17
	s.srcTermID = tool.TrimStr(frame[index : index+21])
	index += 21
	s.chargeTermID = tool.TrimStr(frame[index : index+21])
	index += 21
	s.destTermIDCount = frame[index]
	index++
	for i := byte(0); i < s.destTermIDCount; i++ {
		s.destTermID = append(s.destTermID, tool.TrimStr(frame[index:index+21]))
		index += 21
	}
	s.msgLength = frame[index]
	index++
	content := frame[index : index+int(s.msgLength)]
	s.msgBytes = content
	if content[0] == 0x05 && content[1] == 0x00 && content[2] == 0x03 {
		content = content[6:]
	}
	index += int(s.msgLength)
	tmp, _ := GbDecoder.Bytes(content)
	s.msgContent = string(tmp)
	s.reserve = tool.TrimStr(frame[index : index+8])
	index += 8
	buf := bytes.NewBuffer(frame[index:])
	s.tlvList, _ = tool.Read(buf)
	return nil
}

func (s *Submit) ToResponse(code uint32) interface{} {
	header := *s.MessageHeader
	header.RequestId = CmdSubmitResp
	header.PacketLength = 26
	resp := &SubmitResp{MessageHeader: &header}
	resp.status = code
	resp.msgId = MsgIdSeq.NextSeq()
	return resp
}

func (s *Submit) String() string {
	return fmt.Sprintf("{ header: %v, msgType: %v, isReport: %v, priority: %v, serviceID: %v, "+
		"feeType: %v, feeCode: %v, fixedFee: %v, msgFormat: %v, atTime: %v, validTime: %v, srcTermID: %v, "+
		"chargeTermID: %v, destTermIDCount: %v, destTermID: %v, msgLength: %v, msgContent: %#x..., "+
		"reserve: %v, tlvList: %s }",
		s.MessageHeader, s.msgType, s.needReport, s.priority, s.serviceID,
		s.feeType, s.feeCode, s.fixedFee, s.msgFormat, s.atTime, s.validTime, s.srcTermID,
		s.chargeTermID, s.destTermIDCount, s.destTermID, s.msgLength, s.msgBytes[:6],
		s.reserve, s.tlvList)
}

func (r *SubmitResp) Encode() []byte {
	frame := r.MessageHeader.Encode()
	index := 12
	copy(frame[index:index+10], r.msgId)
	index += 10
	binary.BigEndian.PutUint32(frame[index:index+4], r.status)
	return frame
}

func (r *SubmitResp) Decode(header *MessageHeader, frame []byte) error {
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

func (r *SubmitResp) String() string {
	return fmt.Sprintf("{ header: %s, msgId: %x, status: {%d:%s} }", r.MessageHeader, r.msgId, r.status, StatMap[r.status])
}

func (r *SubmitResp) MsgId() string {
	return fmt.Sprintf("%x", r.msgId)
}
