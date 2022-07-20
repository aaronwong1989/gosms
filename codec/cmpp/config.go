package cmpp

import (
	"errors"

	"github.com/aaronwong1989/gosms/codec"
	"github.com/aaronwong1989/gosms/comm/logging"
	"github.com/aaronwong1989/gosms/comm/yml_config"
)

var log = logging.GetDefaultLogger()
var ErrorPacket = errors.New("error packet")
var Conf yml_config.YmlConfig
var Seq32 codec.Sequence32
var Seq64 codec.Sequence64

// type Config struct {
// 	// 公共参数
// 	SourceAddr         string        `yaml:"source-addr"`
// 	SharedSecret       string        `yaml:"shared-secret"`
// 	AuthCheck          bool          `yaml:"auth-check"`
// 	Version            uint8         `yaml:"version"`
// 	MaxCons            int           `yaml:"max-cons"`
// 	ActiveTestDuration time.Duration `yaml:"active-test-duration"`
// 	DataCenterId       int32         `yaml:"datacenter-id"`
// 	WorkerId           int32         `yaml:"worker-id"`
// 	ReceiveWindowSize  int           `yaml:"receive-window-size"`
// 	MaxPoolSize        int           `yaml:"max-pool-size"`
//
// 	// MT消息相关
// 	RegisteredDel   uint8         `yaml:"need-report"`
// 	MsgLevel        uint8         `yaml:"default-msg-level"`
// 	FeeUsertype     uint8         `yaml:"fee-user-type"`
// 	FeeTerminalType uint8         `yaml:"c"`
// 	SrcId           string        `yaml:"sms-display-no"`
// 	ServiceId       string        `yaml:"service-id"`
// 	FeeTerminalId   string        `yaml:"fee-terminal-id"`
// 	FeeType         string        `yaml:"fee-type"`
// 	FeeCode         string        `yaml:"fee-code"`
// 	LinkID          string        `yaml:"link-id"`
// 	ValidDuration   time.Duration `yaml:"default-valid-duration"`
//
// 	// 模拟网关相关参数
// 	SuccessRate     int32 `yaml:"success-rate"`
// 	MinSubmitRespMs int32 `yaml:"min-submit-resp-ms"`
// 	MaxSubmitRespMs int32 `yaml:"max-submit-resp-ms"`
// 	FixReportRespMs int32 `yaml:"fix-report-resp-ms"`
// }
