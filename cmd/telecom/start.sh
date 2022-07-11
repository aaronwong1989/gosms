#!/bin/sh

pkill telecom
pkill telecom

# -1=debug, 0=info, 1=warn..., default to info
export GNET_LOGGING_LEVEL=0
export GNET_LOGGING_FILE="/Users/huangzhonghui/logs/telecom.log"
export TELECOM_CONF_PATH="/Users/huangzhonghui/.telecom.yaml"

mkdir -p /Users/huangzhonghui/logs

# optional args --port 1234 --multicore=false
# default  args --port 9100 --multicore=true
nohup ./telecom --port 9100 --multicore=true >panic.log 2>&1 &

sleep 3
tail -10 /Users/huangzhonghui/logs/telecom.log
sleep 7
top -pid "$(cat telecom.pid)"
