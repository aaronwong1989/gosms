package cmpp

import (
	"errors"
	"io/ioutil"
	"os"
	"time"

	"gopkg.in/yaml.v3"

	"gosms/comm/logging"
)

var log = logging.GetDefaultLogger()
var ErrorPacket = errors.New("error packet")
var Conf = NewConfig()
var Seq32 Sequence32
var Seq64 Sequence64

type Config struct {
	// 公共参数
	SourceAddr         string        `yaml:"source-addr"`
	SharedSecret       string        `yaml:"shared-secret"`
	AuthCheck          bool          `yaml:"auth-check"`
	Version            uint8         `yaml:"version"`
	MaxCons            int           `yaml:"max-cons"`
	ActiveTestDuration time.Duration `yaml:"active-test-duration"`
	DataCenterId       int32         `yaml:"datacenter-id"`
	WorkerId           int32         `yaml:"worker-id"`
	ReceiveWindowSize  int           `yaml:"receive-window-size"`
	MaxPoolSize        int           `yaml:"max-pool-size"`

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

func NewConfig() *Config {
	var conf Config
	path := os.Getenv("ISMG_CONF_PATH")
	if len(path) == 0 {
		// TODO define your default fallback path
		path = "/Users/huangzhonghui/GolandProjects/gosms/cmd/ismg/cmpp/cmpp.yaml"
	}
	log.Infof("[Conf     ] path=%s", path)
	config, err := ioutil.ReadFile(path)
	if err != nil {
		panic(err)
	}
	err = yaml.Unmarshal(config, &conf)
	log.Infof("[Conf     ] %+v", conf)
	if err != nil {
		panic(err)
	}
	return &conf
}
