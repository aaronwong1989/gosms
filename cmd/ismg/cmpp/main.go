package main

import (
	"math/rand"
	"time"

	"github.com/aaronwong1989/yaml_config"

	"github.com/aaronwong1989/gosms/codec/cmpp"
	"github.com/aaronwong1989/gosms/comm"
	"github.com/aaronwong1989/gosms/comm/logging"
	"github.com/aaronwong1989/gosms/comm/snowflake"
)

var log = logging.GetDefaultLogger()

func main() {
	rand.Seed(time.Now().Unix()) // 随机种子
	cmpp.Conf = yaml_config.CreateYamlFactory("config", "cmpp", "gosms")
	cmpp.Conf.ConfigFileChangeListen()

	dc := cmpp.Conf.GetInt("data-center-id")
	wk := cmpp.Conf.GetInt("worker-id")
	cmpp.Seq32 = comm.NewCycleSequence(int32(dc), int32(wk))
	cmpp.Seq64 = snowflake.NewSnowflake(int64(dc), int64(wk))
	cmpp.ReportSeq = comm.NewCycleSequence(int32(dc), int32(wk))
	StartServer()
}
