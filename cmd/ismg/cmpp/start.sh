#!/bin/sh

pkill cmpp.ismg
pkill cmpp.ismg

# -1=debug, 0=info, 1=warn..., default to info
export GNET_LOGGING_LEVEL=0
export GNET_LOGGING_FILE="/Users/huangzhonghui/logs/cmpp.log"
export ISMG_CONF_PATH="/Users/huangzhonghui/.cmpp.yaml"

mkdir -p /Users/huangzhonghui/logs

# optional args --port 1234 --multicore=false
# default  args --port 9000 --multicore=true
nohup ./cmpp.ismg --port 9000 --multicore=true >panic.log 2>&1 &

sleep 3
tail -10 /Users/huangzhonghui/logs/cmpp.log
sleep 7
top -pid "$(cat cmpp.pid)"
