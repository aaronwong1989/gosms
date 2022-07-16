package main

import (
	"math/rand"
	"time"

	"gosms/codec/smgp"
	"gosms/comm"
	"gosms/comm/logging"
)

var log = logging.GetDefaultLogger()

func main() {
	rand.Seed(time.Now().Unix()) // 随机种子
	smgp.Seq32 = comm.NewCycleSequence(smgp.Conf.DataCenterId, smgp.Conf.WorkerId)
	smgp.Seq80 = comm.NewBcdSequence(smgp.Conf.SmgwId)

	StartServer()
}
