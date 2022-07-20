package main

import (
	"math/rand"
	"time"

	"github.com/aaronwong1989/gosms/codec/smgp"
	"github.com/aaronwong1989/gosms/comm"
	"github.com/aaronwong1989/gosms/comm/logging"
	"github.com/aaronwong1989/gosms/comm/yml_config"
)

var log = logging.GetDefaultLogger()

func main() {
	rand.Seed(time.Now().Unix()) // 随机种子

	smgp.Conf = yml_config.CreateYamlFactory("smgp")
	smgp.Conf.ConfigFileChangeListen()

	dc := smgp.Conf.GetInt("data-center-id")
	wk := smgp.Conf.GetInt("worker-id")
	smgwId := smgp.Conf.GetString("smgw-id")
	smgp.Seq32 = comm.NewCycleSequence(int32(dc), int32(wk))
	smgp.Seq80 = comm.NewBcdSequence(smgwId)
	StartServer()
}
