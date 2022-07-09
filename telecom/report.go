package telecom

import (
	"fmt"
	"strings"
	"time"
)

type Report struct {
	id         string // 【10字节】状态报告对应原短消息的MsgID
	sub        string // 【3字节】取缺省值001
	dlvrd      string // 【3字节】取缺省值001
	submitDate string // 【10字节】短消息提交时间（格式：年年月月日日时时分分，例如010331200000）
	doneDate   string // 【10字节】短消息提交时间（格式：年年月月日日时时分分，例如010331200000）
	stat       string // 【7字节】短消息的最终状态
	err        string // 【3字节】短消息的最终状态
	txt        string // 【20字节】前3个字节，表示短消息长度（用ASCII码表示），后17个字节表示短消息的内容
}

func NewReport(id string) *Report {
	report := &Report{id: id}
	report.sub = "001"
	report.dlvrd = "001"
	report.submitDate = time.Now().Format("0601021504")
	report.doneDate = time.Now().Add(time.Minute).Format("0601021504")
	report.txt = "000"
	// 判断序号的时间戳部分
	switch time.Now().Unix() % 1000 {
	case 1:
		report.err = "001"
		report.stat = reportStatMap["001"]
	case 2:
		report.err = "002"
		report.stat = reportStatMap["002"]
	case 3:
		report.err = "003"
		report.stat = reportStatMap["003"]
	case 4:
		report.err = "004"
		report.stat = reportStatMap["004"]
	case 5:
		report.err = "005"
		report.stat = reportStatMap["005"]
	case 6:
		report.err = "006"
		report.stat = reportStatMap["006"]
	case 7:
		report.err = "007"
		report.stat = reportStatMap["007"]
	case 8:
		report.err = "008"
		report.stat = reportStatMap["008"]
	case 9:
		report.err = "009"
		report.stat = reportStatMap["009"]
	case 10:
		report.err = "010"
		report.stat = reportStatMap["010"]
	default:
		report.err = "000"
		report.stat = reportStatMap["000"]
	}
	return report
}

func (rt *Report) String() string {
	return fmt.Sprintf("id:%s sub:%s dlvrd:%s Submit_date:%s done_date:%s stat:%s err:%s Text:%s",
		rt.id, rt.sub, rt.dlvrd, rt.submitDate, rt.doneDate, rt.stat, rt.err, rt.txt)
}

func (rt *Report) Encode() string {
	return rt.String()
}

func (rt *Report) Decode(s string) error {
	ss := strings.Split(s, " ")
	if len(ss) < 8 {
		return ErrorPacket
	} else {
		// id
		sss := strings.Split(ss[0], ":")
		if len(sss) != 2 {
			return ErrorPacket
		}
		rt.id = sss[1]
		// sub
		sss = strings.Split(ss[1], ":")
		if len(sss) != 2 {
			return ErrorPacket
		}
		rt.sub = sss[1]
		// dlvrd
		sss = strings.Split(ss[2], ":")
		if len(sss) != 2 {
			return ErrorPacket
		}
		rt.dlvrd = sss[1]
		// Submit_date
		sss = strings.Split(ss[3], ":")
		if len(sss) != 2 {
			return ErrorPacket
		}
		rt.submitDate = sss[1]
		// done_date
		sss = strings.Split(ss[4], ":")
		if len(sss) != 2 {
			return ErrorPacket
		}
		rt.doneDate = sss[1]
		// stat
		sss = strings.Split(ss[5], ":")
		if len(sss) != 2 {
			return ErrorPacket
		}
		rt.stat = sss[1]
		// err
		sss = strings.Split(ss[6], ":")
		if len(sss) != 2 {
			return ErrorPacket
		}
		rt.err = sss[1]
		// Text
		sss = strings.Split(ss[7], ":")
		if len(sss) != 2 {
			return ErrorPacket
		}
		rt.txt = sss[1]
	}
	return nil
}

var reportStatMap = map[string]string{
	"000": "DELIVRD", // 成功
	"001": "EXPIRED", // 用户不能通信
	"002": "EXPIRED", // 用户忙
	"003": "UNDELIV", // 终端无此部件号
	"004": "UNDELIV", // 非法用户
	"005": "UNDELIV", // 用户在黑名单内
	"006": "UNDELIV", // 系统错误
	"007": "EXPIRED", // 用户内存满
	"008": "UNDELIV", // 非信息终端
	"009": "UNDELIV", // 数据错误
	"010": "UNDELIV", // 数据丢失
	"999": "UNKNOWN", // 未知错误
}
