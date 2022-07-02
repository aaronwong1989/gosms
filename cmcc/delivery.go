package cmcc

import (
	"encoding/binary"
	"fmt"
)

// Delivery 上行短信或状态报告，不支持长短信
type Delivery struct {
	*MessageHeader

	msgId              uint64  // 信息标识
	destId             string  // 目的号码 21
	serviceId          string  // 业务标识，是数字、字母和符号的组合。 10
	tpPid              uint8   // 见Submit
	tpUdhi             uint8   // 见Submit
	msgFmt             uint8   // 见Submit
	srcTerminalId      string  // 源终端MSISDN号码（状态报告时填为CMPP_SUBMIT消息的目的终端号码）
	srcTerminalType    uint8   // 源终端号码类型，0：真实号码；1：伪码
	registeredDelivery uint8   // 是否为状态报告
	msgLength          uint8   // 消息长度
	msgContent         string  // 非状态报告的消息内容
	report             *Report // 状态报告的消息内容
	linkID             string  // 点播业务使用的LinkID，非点播类业务的MT流程不使用该字段
}

func NewDelivery(phone string, msg string, dest string, serviceId string) *Delivery {
	dly := &Delivery{}
	dly.srcTerminalId = phone
	dly.srcTerminalType = 0
	setMsgContent(dly, msg)

	if dest != "" {
		dly.destId = dest
	} else {
		dly.destId = Conf.SrcId
	}
	if serviceId != "" {
		dly.serviceId = serviceId
	} else {
		dly.serviceId = Conf.ServiceId
	}

	header := MessageHeader{
		TotalLength: 109 + uint32(dly.msgLength),
		CommandId:   CMPP_DELIVER,
		SequenceId:  uint32(Sequence32.NextVal())}
	dly.MessageHeader = &header
	return dly
}

func (d *Delivery) Encode() []byte {
	frame := d.MessageHeader.Encode()
	binary.BigEndian.PutUint64(frame[12:20], d.msgId)
	copy(frame[20:41], d.destId)
	copy(frame[41:51], d.serviceId)
	frame[51] = d.tpPid
	frame[52] = d.tpUdhi
	frame[53] = d.msgFmt
	copy(frame[54:86], d.srcTerminalId)
	frame[86] = d.srcTerminalType
	frame[87] = d.registeredDelivery
	frame[88] = d.msgLength
	index := 89 + int(d.msgLength)
	if d.registeredDelivery == 1 {
		// 状态报告
		copy(frame[89:index], d.report.Encode())
	} else {
		// 上行短信，不支持长短信，固定选用第一片 （New时需处理）
		slices := MsgSlices(d.msgFmt, d.msgContent)
		// 不支持长短信，固定选用第一片
		content := slices[0]
		copy(frame[89:index], content)
	}
	copy(frame[index:index+20], d.linkID)

	return frame
}

func (d *Delivery) Decode(header *MessageHeader, frame []byte) error {
	if header == nil || header.CommandId != CMPP_DELIVER || uint32(len(frame)) < (header.TotalLength-HEAD_LENGTH) {
		return ErrorPacket
	}
	d.MessageHeader = header
	d.msgId = binary.BigEndian.Uint64(frame[0:8])
	d.destId = TrimStr(frame[8:29])
	d.destId = TrimStr(frame[29:39])
	d.tpPid = frame[39]
	d.tpUdhi = frame[40]
	d.msgFmt = frame[41]
	d.srcTerminalId = TrimStr(frame[42:74])
	d.srcTerminalType = frame[74]
	d.registeredDelivery = frame[75]

	d.msgLength = frame[76]
	index := 77 + int(d.msgLength)
	if d.registeredDelivery == 1 {
		rpt := &Report{}
		err := rpt.Decode(frame[77:index])
		if err != nil {
			return err
		}
		d.report = rpt
	} else {
		d.msgContent = TrimStr(frame[77:index])
	}
	d.linkID = TrimStr(frame[index:])
	return nil
}

func (d *Delivery) String() string {
	content := d.msgContent
	if d.registeredDelivery == 1 {
		content = d.report.String()
	}

	return fmt.Sprintf("{ msgId: %d, destId: %v, serviceId: %v, tpPid: %d, tpUdhi: %d, msgFmt: %d, "+
		"srcTerminalId: %v, srcTerminalType: %d, registeredDelivery: %d, "+
		"msgLength: %d, setMsgContent: %v, linkID: %v }",
		d.msgId, d.destId, d.serviceId, d.tpPid, d.tpUdhi, d.msgFmt,
		d.srcTerminalId, d.srcTerminalType, d.registeredDelivery,
		d.msgLength, content, d.linkID,
	)
}

func setMsgContent(dly *Delivery, msg string) {
	dly.msgFmt = MsgFmt(msg)
	var l int
	if dly.msgFmt == 8 {
		l = 2 * len([]rune(msg))
		if l > 140 {
			// 只取前70个字符
			rs := []rune(msg)
			msg = string(rs[:70])
			l = 140
		}
	} else {
		l = len(msg)
		if l > 160 {
			// 只取前160个字符
			msg = msg[:160]
			l = 160
		}
	}
	dly.msgLength = uint8(l)
	dly.msgContent = msg
}
