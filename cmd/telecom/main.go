package main

import (
	"math/rand"
	"time"

	"sms-vgateway/comm"
	"sms-vgateway/telecom"
)

func main() {
	rand.Seed(time.Now().Unix()) // 随机种子
	telecom.RequestSeq = comm.NewCycleSequence(telecom.Conf.DataCenterId, telecom.Conf.WorkerId)
	telecom.MsgIdSeq = comm.NewBcdSequence(telecom.Conf.SmgwId)

	// IO密集型程序可调大此值，比真实处理器更多
	// runtime.GOMAXPROCS(256)

	telecom.StartServer()
}
