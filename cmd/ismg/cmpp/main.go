package main

import (
	"math/rand"
	"time"

	"gosms/codec/cmpp"
	"gosms/comm"
	"gosms/comm/logging"
	"gosms/comm/snowflake"
)

var log = logging.GetDefaultLogger()

func main() {
	rand.Seed(time.Now().Unix()) // 随机种子
	cmpp.Seq32 = comm.NewCycleSequence(cmpp.Conf.DataCenterId, cmpp.Conf.WorkerId)
	cmpp.Seq64 = snowflake.NewSnowflake(int64(cmpp.Conf.DataCenterId), int64(cmpp.Conf.WorkerId))
	StartServer()
}
