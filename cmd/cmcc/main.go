package main

import (
	"math/rand"
	"time"

	"sms-vgateway/cmcc"
	"sms-vgateway/comm"
	"sms-vgateway/comm/snowflake"
)

func main() {
	rand.Seed(time.Now().Unix()) // 随机种子
	cmcc.RequestSeq = comm.NewCycleSequence(cmcc.Conf.DataCenterId, cmcc.Conf.WorkerId)
	cmcc.MsgIdSeq = snowflake.NewSnowflake(int64(cmcc.Conf.DataCenterId), int64(cmcc.Conf.WorkerId))
	cmcc.StartServer()
}
