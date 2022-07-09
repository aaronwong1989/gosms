package telecom

import (
	"time"

	"sms-vgateway/comm"
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
	s.needReport = Conf.NeedReport
	if options.needReport != Conf.NeedReport {
		s.needReport = options.needReport
	}

	s.priority = Conf.Priority
	if options.priority != Conf.Priority {
		s.priority = options.priority
	}

	s.serviceID = Conf.ServiceId
	if options.serviceID != Conf.ServiceId {
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
		vt.Add(Conf.ValidDuration)
	}
	s.validTime = comm.FormatTime(vt)

	s.srcTermID = Conf.DisplayNo
	if options.srcTermID != "" {
		s.srcTermID += options.srcTermID
	}
}
