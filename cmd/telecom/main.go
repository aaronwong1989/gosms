package main

import (
	"sms-vgateway/snowflake32"
	"sms-vgateway/telecom"
)

func main() {
	telecom.Sequence32 = snowflake32.NewSnowflake(telecom.Conf.DataCenterId, telecom.Conf.WorkerId)
	telecom.MsgIdSeq = snowflake32.NewTelecomflake(telecom.Conf.SmgwId)
	telecom.StartServer()
}
