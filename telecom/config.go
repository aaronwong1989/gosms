package telecom

import (
	"errors"
	"io/ioutil"
	"os"
	"time"

	"golang.org/x/text/encoding/simplifiedchinese"
	"gopkg.in/yaml.v3"

	"sms-vgateway/comm"
	"sms-vgateway/comm/logging"
)

var log = logging.GetDefaultLogger()
var Conf Config
var ErrorPacket = errors.New("error packet")
var GbEncoder = simplifiedchinese.GB18030.NewEncoder()
var GbDecoder = simplifiedchinese.GB18030.NewDecoder()
var RequestSeq = comm.NewCycleSequence(Conf.DataCenterId, Conf.WorkerId)
var MsgIdSeq = comm.NewBcdSequence(Conf.SmgwId)

type Config struct {
	// 公共参数
	ClientId           string        `yaml:"client-id"`
	SharedSecret       string        `yaml:"shared-secret"`
	AuthCheck          bool          `yaml:"auth-check"`
	Version            byte          `yaml:"version"`
	MaxCons            int           `yaml:"max-cons"`
	ActiveTestDuration time.Duration `yaml:"active-test-duration"`
	DataCenterId       int32         `yaml:"datacenter-id"`
	WorkerId           int32         `yaml:"worker-id"`
	SmgwId             string        `yaml:"smgw-id"`
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

const (
	TP_pid           = uint16(0x0001)
	TP_udhi          = uint16(0x0002)
	LinkID           = uint16(0x0003)
	ChargeUserType   = uint16(0x0004)
	ChargeTermType   = uint16(0x0005)
	ChargeTermPseudo = uint16(0x0006)
	DestTermType     = uint16(0x0007)
	DestTermPseudo   = uint16(0x0008)
	PkTotal          = uint16(0x0009)
	PkNumber         = uint16(0x000A)
	SubmitMsgType    = uint16(0x000B)
	SPDealReslt      = uint16(0x000C)
	SrcTermType      = uint16(0x000D)
	SrcTermPseudo    = uint16(0x000E)
	NodesCount       = uint16(0x000F)
	MsgSrc           = uint16(0x0010)
	SrcType          = uint16(0x0011)
	MServiceID       = uint16(0x0012)
)

var StatMap = map[uint32]string{
	0:  "成功",
	1:  "系统忙",
	2:  "超过最大连接数",
	10: "消息结构错",
	11: "命令字错",
	12: "序列号重复",
	20: "IP地址错",
	21: "认证错",
	22: "版本太高",
	30: "非法消息类型（MsgType）",
	31: "非法优先级（Priority）",
	32: "非法资费类型（FeeType）",
	33: "非法资费代码（FeeCode）",
	34: "非法短消息格式（MsgFormat）",
	35: "非法时间格式",
	36: "非法短消息长度（MsgLength）",
	37: "有效期已过",
	38: "非法查询类别（QueryType）",
	39: "路由错误",
	40: "非法包月费/封顶费（FixedFee）",
	41: "非法更新类型（UpdateType）",
	42: "非法路由编号（RouteId）",
	43: "非法服务代码（ServiceId）",
	44: "非法有效期（ValidTime）",
	45: "非法定时发送时间（AtTime）",
	46: "非法发送用户号码（SrcTermId）",
	47: "非法接收用户号码（DestTermId）",
	48: "非法计费用户号码（ChargeTermId）",
	49: "非法SP服务代码（SPCode）",
	56: "非法源网关代码（SrcGatewayID）",
	57: "非法查询号码（QueryTermID）",
	58: "没有匹配路由",
	59: "非法SP类型（SPType）",
	60: "非法上一条路由编号（LastRouteID）",
	61: "非法路由类型（RouteType）",
	62: "非法目标网关代码（DestGatewayID）",
	63: "非法目标网关IP（DestGatewayIP）",
	64: "非法目标网关端口（DestGatewayPort）",
	65: "非法路由号码段（TermRangeID）",
	66: "非法终端所属省代码（ProvinceCode）",
	67: "非法用户类型（UserType）",
	68: "本节点不支持路由更新",
	69: "非法SP企业代码（SPID）",
	70: "非法SP接入类型（SPAccessType）",
	71: "路由信息更新失败",
	72: "非法时间戳（Time）",
	73: "非法业务代码（MServiceID）",
	74: "SP禁止下发时段",
	75: "SP发送超过日流量",
	76: "SP帐号过有效期",
}
