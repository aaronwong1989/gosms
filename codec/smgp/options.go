package smgp

import (
	"time"

	"github.com/aaronwong1989/gosms/comm"
)

type MtOptions struct {
	needReport    byte          // SP是否要求返回状态报告
	priority      byte          // 短消息发送优先级,0-3
	serviceID     string        // 业务代码
	atTime        time.Time     // 短消息定时发送时间
	validDuration time.Duration // 短消息有效时长
	srcTermID     string        // 会拼接到配置文件的sms-display-no后面
}

func (s *Submit) SetOptions(options MtOptions) {
	s.needReport = byte(Conf.GetInt("need-report"))
	// 有点小bug，不能通过传参的方式设置未变量的"零值"
	if options.needReport != 0 {
		s.needReport = options.needReport
	}

	s.priority = byte(Conf.GetInt("priority"))
	// 有点小bug，不能通过传参的方式设置未变量的"零值"
	if options.priority != 0 {
		s.priority = options.priority
	}

	s.serviceID = Conf.GetString("service-id")
	if options.serviceID != "" {
		s.serviceID = options.serviceID
	}

	if options.atTime.Year() != 1 {
		s.atTime = comm.FormatTime(options.atTime)
	} else {
		s.atTime = comm.FormatTime(time.Now())
	}

	vt := time.Now()
	if options.validDuration != 0 {
		vt.Add(options.validDuration)
	} else {
		vt.Add(Conf.GetDuration("default-valid-duration"))
	}
	s.validTime = comm.FormatTime(vt)

	s.srcTermID = Conf.GetString("sms-display-no")
	if options.srcTermID != "" {
		s.srcTermID += options.srcTermID
	}
}
