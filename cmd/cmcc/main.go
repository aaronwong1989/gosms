package main

import (
	"sms-vgateway/cmcc"
	"sms-vgateway/snowflake"
	"sms-vgateway/snowflake32"
)

func main() {
	cmcc.Sequence32 = snowflake32.NewSnowflake(cmcc.Conf.DataCenterId, cmcc.Conf.WorkerId)
	cmcc.Sequence64 = snowflake.NewSnowflake(int64(cmcc.Conf.DataCenterId), int64(cmcc.Conf.WorkerId))
	cmcc.StartServer()
}
