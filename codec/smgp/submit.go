package smgp

import (
	"bytes"
	"encoding/binary"
	"fmt"

	"github.com/aaronwong1989/gosms/comm"
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
	validTime       string        // 【17字节】短消息有效时间
	atTime          string        // 【17字节】短消息定时发送时间
	srcTermID       string        // 【21字节】短信息发送方号码
	chargeTermID    string        // 【21字节】计费用户号码
	destTermIDCount byte          // 【1字节】短消息接收号码总数
	destTermID      []string      // 【21*DestTermCount字节】短消息接收号码
	msgLength       byte          // 【1字节】短消息长度
	msgContent      string        // 【MsgLength字节】短消息内容
	msgBytes        []byte        // 消息内容按照Msg_Fmt编码后的数据
	reserve         string        // 【8字节】保留
	tlvList         *comm.TlvList // 【TLV】可选项参数
}

type SubmitResp struct {
	*MessageHeader
	msgId  []byte // 【10字节】短消息流水号
	status uint32
}

const MtBaseLen = 126

func NewSubmit(phones []string, content string, options MtOptions) (messages []*Submit) {

	head := &MessageHeader{PacketLength: MtBaseLen, RequestId: CmdSubmit, SequenceId: uint32(Seq32.NextVal())}
	mt := &Submit{}
	mt.MessageHeader = head
	mt.SetOptions(options)
	mt.msgType = 6
	// 从配置文件设置属性
	mt.feeType = Conf.GetString("fee-type")
	mt.feeCode = Conf.GetString("fee-code")
	mt.chargeTermID = Conf.GetString("charge-term-id")
	mt.fixedFee = Conf.GetString("fixed-fee")
	// 初步设置入参
	mt.destTermID = phones
	mt.destTermIDCount = byte(len(phones))

	mt.msgFormat = 15
	data, err := GbEncoder.Bytes([]byte(content))
	if err != nil {
		return nil
	}
	slices := comm.ToTPUDHISlices(data, 140)
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
				sub.SequenceId = uint32(Seq32.NextVal())
			}
			sub.msgLength = byte(len(dt))
			sub.msgBytes = dt
			l := 0
			sub.tlvList = comm.NewTlvList()
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
	index = comm.CopyByte(frame, s.msgType, index)
	index = comm.CopyByte(frame, s.needReport, index)
	index = comm.CopyByte(frame, s.priority, index)
	index = comm.CopyStr(frame, s.serviceID, index, 10)
	index = comm.CopyStr(frame, s.feeType, index, 2)
	index = comm.CopyStr(frame, s.feeCode, index, 6)
	index = comm.CopyStr(frame, s.fixedFee, index, 6)
	index = comm.CopyByte(frame, s.msgFormat, index)
	index = comm.CopyStr(frame, s.validTime, index, 17)
	index = comm.CopyStr(frame, s.atTime, index, 17)
	index = comm.CopyStr(frame, s.srcTermID, index, 21)
	index = comm.CopyStr(frame, s.chargeTermID, index, 21)
	index = comm.CopyByte(frame, s.destTermIDCount, index)
	for _, tid := range s.destTermID {
		index = comm.CopyStr(frame, tid, index, 21)
	}

	index = comm.CopyByte(frame, s.msgLength, index)
	copy(frame[index:index+int(s.msgLength)], s.msgBytes)
	index += +int(s.msgLength)
	index = comm.CopyStr(frame, s.reserve, index, 8)
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
	s.serviceID = comm.TrimStr(frame[index : index+10])
	index += 10
	s.feeType = comm.TrimStr(frame[index : index+2])
	index += 2
	s.feeCode = comm.TrimStr(frame[index : index+6])
	index += 6
	s.fixedFee = comm.TrimStr(frame[index : index+6])
	index += 6
	s.msgFormat = frame[index]
	index++
	s.validTime = comm.TrimStr(frame[index : index+17])
	index += 17
	s.atTime = comm.TrimStr(frame[index : index+17])
	index += 17
	s.srcTermID = comm.TrimStr(frame[index : index+21])
	index += 21
	s.chargeTermID = comm.TrimStr(frame[index : index+21])
	index += 21
	s.destTermIDCount = frame[index]
	index++
	for i := byte(0); i < s.destTermIDCount; i++ {
		s.destTermID = append(s.destTermID, comm.TrimStr(frame[index:index+21]))
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
	s.reserve = comm.TrimStr(frame[index : index+8])
	index += 8
	// 一个tlv至少5字节
	if uint32(index+5) < s.PacketLength {
		buf := bytes.NewBuffer(frame[index:])
		s.tlvList, _ = comm.Read(buf)
	}
	return nil
}

func (s *Submit) ToResponse(code uint32) interface{} {
	header := *s.MessageHeader
	header.RequestId = CmdSubmitResp
	header.PacketLength = 26
	resp := &SubmitResp{MessageHeader: &header}
	resp.status = code
	resp.msgId = Seq80.NextVal()
	return resp
}

func (s *Submit) String() string {
	bts := s.msgBytes
	if s.msgLength > 6 {
		bts = s.msgBytes[:6]
	}
	return fmt.Sprintf("{ header: %v, msgType: %v, NeedReport: %v, Priority: %v, ServiceID: %v, "+
		"feeType: %v, feeCode: %v, fixedFee: %v, msgFormat: %v, validTime: %v, AtTime: %v, SrcTermID: %v, "+
		"chargeTermID: %v, destTermIDCount: %v, destTermID: %v, msgLength: %v, msgContent: %#x..., "+
		"reserve: %v, tlvList: %s }",
		s.MessageHeader, s.msgType, s.needReport, s.priority, s.serviceID,
		s.feeType, s.feeCode, s.fixedFee, s.msgFormat, s.validTime, s.atTime, s.srcTermID,
		s.chargeTermID, s.destTermIDCount, s.destTermID, s.msgLength, bts,
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

func (r *SubmitResp) MsgId() []byte {
	return r.msgId
}

func (r *SubmitResp) Status() uint32 {
	return r.status
}
