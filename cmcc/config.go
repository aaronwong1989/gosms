package cmcc

import (
	"io/ioutil"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

var Conf Config

type Config struct {
	// 公共参数
	SourceAddr         string        `yaml:"sp-id"`
	SharedSecret       string        `yaml:"shared-secret"`
	Version            uint8         `yaml:"version"`
	MaxCons            int           `yaml:"max-cons"`
	ActiveTestDuration time.Duration `yaml:"active-test-duration"`
	DataCenterId       int32         `yaml:"datacenter-id"`
	WorkerId           int32         `yaml:"worker-id"`
	ReceiveWindowSize  int           `yaml:"receive-window-size"`

	// MT消息相关
	RegisteredDel   uint8         `yaml:"need-report"`
	MsgLevel        uint8         `yaml:"default-msg-level"`
	FeeUsertype     uint8         `yaml:"fee-user-type"`
	FeeTerminalType uint8         `yaml:"fee-terminal-type"`
	SrcId           string        `yaml:"sms-display-no"`
	ServiceId       string        `yaml:"service-id"`
	FeeTerminalId   string        `yaml:"fee-terminal-id"`
	FeeType         string        `yaml:"fee-type"`
	FeeCode         string        `yaml:"fee-code"`
	LinkID          string        `yaml:"link-id"`
	ValidDuration   time.Duration `yaml:"default-valid-duration"`

	// 模拟网关相关参数
	SuccessRate     int32 `yaml:"success-rate"`
	MinSubmitRespMs int32 `yaml:"min-submit-resp-ms"`
	MaxSubmitRespMs int32 `yaml:"max-submit-resp-ms"`
	FixReportRespMs int32 `yaml:"fix-report-resp-ms"`
}

func init() {
	path := os.Getenv("CMCC_CONF_PATH")
	if len(path) == 0 {
		path = "/Users/huangzhonghui/.cmcc.yaml"
	}
	config, err := ioutil.ReadFile(path)
	if err != nil {
		panic(err)
		return
	}
	err = yaml.Unmarshal(config, &Conf)
	if err != nil {
		panic(err)
		return
	}
}
