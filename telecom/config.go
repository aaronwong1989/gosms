package telecom

import (
	"errors"
	"io/ioutil"
	"os"
	"time"

	"golang.org/x/text/encoding/simplifiedchinese"
	"gopkg.in/yaml.v3"

	"sms-vgateway/logging"
	"sms-vgateway/snowflake"
	"sms-vgateway/snowflake32"
)

var Conf Config

type Config struct {
	// 公共参数
	clientId           string        `yaml:"client-id"`
	SharedSecret       string        `yaml:"shared-secret"`
	AuthCheck          bool          `yaml:"auth-check"`
	Version            byte          `yaml:"version"`
	MaxCons            int           `yaml:"max-cons"`
	ActiveTestDuration time.Duration `yaml:"active-test-duration"`
	DataCenterId       int32         `yaml:"datacenter-id"`
	WorkerId           int32         `yaml:"worker-id"`
	ReceiveWindowSize  int           `yaml:"receive-window-size"`
	MaxPoolSize        int           `yaml:"max-pool-size"`

	// MT消息相关
	NeedReport    byte          `yaml:"need-report"`
	Priority      byte          `yaml:"priority "`
	DisplayNo     string        `yaml:"sms-display-no"`
	ServiceId     string        `yaml:"service-id"`
	FeeType       string        `yaml:"fee-type"`
	FeeCode       string        `yaml:"fee-code"`
	ChargeTermID  string        `yaml:"charge-term-id"`
	FixedFee      string        `yaml:"fixed-fee"`
	LinkID        string        `yaml:"link-id"`
	ValidDuration time.Duration `yaml:"default-valid-duration"`

	// 模拟网关相关参数
	SuccessRate     int32 `yaml:"success-rate"`
	MinSubmitRespMs int32 `yaml:"min-submit-resp-ms"`
	MaxSubmitRespMs int32 `yaml:"max-submit-resp-ms"`
	FixReportRespMs int32 `yaml:"fix-report-resp-ms"`
}

func init() {
	path := os.Getenv("TELECOM_CONF_PATH")
	if len(path) == 0 {
		// TODO define your default fallback path
		path = "/Users/huangzhonghui/GolandProjects/sms-vgateway/cmd/telecom/telecom.yaml"
	}
	logging.Infof("[Conf     ] path=%s", path)
	config, err := ioutil.ReadFile(path)
	if err != nil {
		panic(err)
		return
	}
	err = yaml.Unmarshal(config, &Conf)
	log.Infof("[Conf     ] %+v", Conf)
	if err != nil {
		panic(err)
		return
	}
}

var ErrorPacket = errors.New("error packet")
var Sequence32 = snowflake32.NewSnowflake(Conf.DataCenterId, Conf.WorkerId)
var Sequence64 = snowflake.NewSnowflake(int64(Conf.DataCenterId), int64(Conf.WorkerId))
var GbEncoder = simplifiedchinese.GB18030.NewEncoder()
var GbDecoder = simplifiedchinese.GB18030.NewDecoder()
var log = logging.GetDefaultLogger()
