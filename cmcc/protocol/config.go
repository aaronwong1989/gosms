package cmcc

import (
	"io/ioutil"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

var Conf Config

type Config struct {
	SourceAddr         string        `yaml:"sourceAddr"`
	SharedSecret       string        `yaml:"sharedSecret"`
	Version            uint8         `yaml:"version"`
	MaxCons            int           `yaml:"max-cons"`
	ActiveTestDuration time.Duration `yaml:"active-test-duration"`
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
