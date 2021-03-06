package smgp

import (
	"time"

	"github.com/aaronwong1989/gosms/comm"
)

type MtOptions struct {
	NeedReport    byte          // SP是否要求返回状态报告
	Priority      byte          // 短消息发送优先级,0-3
	ServiceID     string        // 业务代码
	AtTime        time.Time     // 短消息定时发送时间
	ValidDuration time.Duration // 短消息有效时长
	SrcTermID     string        // 会拼接到配置文件的sms-display-no后面
}

func (s *Submit) SetOptions(options MtOptions) {
	s.needReport = byte(Conf.GetInt("need-report"))
	// 有点小bug，不能通过传参的方式设置未变量的"零值"
	if options.NeedReport != 0 {
		s.needReport = options.NeedReport
	}

	s.priority = byte(Conf.GetInt("Priority"))
	// 有点小bug，不能通过传参的方式设置未变量的"零值"
	if options.Priority != 0 {
		s.priority = options.Priority
	}

	s.serviceID = Conf.GetString("service-id")
	if options.ServiceID != "" {
		s.serviceID = options.ServiceID
	}

	if options.AtTime.Year() != 1 {
		s.atTime = comm.FormatTime(options.AtTime)
	} else {
		s.atTime = comm.FormatTime(time.Now())
	}

	vt := time.Now()
	if options.ValidDuration != 0 {
		vt.Add(options.ValidDuration)
	} else {
		vt.Add(Conf.GetDuration("default-valid-duration"))
	}
	s.validTime = comm.FormatTime(vt)

	s.srcTermID = Conf.GetString("sms-display-no")
	if options.SrcTermID != "" {
		s.srcTermID += options.SrcTermID
	}
}
