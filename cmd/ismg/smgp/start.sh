#!/bin/sh

pkill smgp.ismg
pkill smgp.ismg

# -1=debug, 0=info, 1=warn..., default to info
export GNET_LOGGING_LEVEL=0
export GNET_LOGGING_FILE="/Users/huangzhonghui/logs/smgp.log"

mkdir -p /Users/huangzhonghui/logs

# optional args --port 1234 --multicore=false
# default  args --port 9100 --multicore=true
nohup ./smgp.ismg --port 9100 --multicore=true >panic.log 2>&1 &

sleep 3
tail -10 /Users/huangzhonghui/logs/smgp.log
sleep 7
top -pid "$(cat smgp.pid)"
