package main

import (
	"math/rand"
	"time"

	"github.com/aaronwong1989/yaml_config"

	"github.com/aaronwong1989/gosms/codec/smgp"
	"github.com/aaronwong1989/gosms/comm"
	"github.com/aaronwong1989/gosms/comm/logging"
)

var log = logging.GetDefaultLogger()

func main() {
	rand.Seed(time.Now().Unix()) // 随机种子

	smgp.Conf = yaml_config.CreateYamlFactory("config", "smgp", "gosms")
	smgp.Conf.ConfigFileChangeListen()

	dc := smgp.Conf.GetInt("data-center-id")
	wk := smgp.Conf.GetInt("worker-id")
	smgwId := smgp.Conf.GetString("smgw-id")
	smgp.Seq32 = comm.NewCycleSequence(int32(dc), int32(wk))
	smgp.Seq80 = comm.NewBcdSequence(smgwId)
	StartServer()
}
