#!/bin/sh

go clean
go mod tidy
# 编译
go build -trimpath
go test -v server_test.go -test.run TestClient -c

mkdir -p            ~/cmcc
mv cmcc             ~/cmcc/
mv main.test        ~/cmcc/
cp start.sh         ~/cmcc/
cp -f cmcc.yaml     ~/.cmcc.yaml

# 拷贝源码
cp -rf ~/GolandProjects/sms-vgateway/cmcc         ~/cmcc/sms-vgateway/
cp -rf ~/GolandProjects/sms-vgateway/telecom      ~/cmcc/sms-vgateway/
cp -rf ~/GolandProjects/sms-vgateway/comm         ~/cmcc/sms-vgateway/
cp -rf ~/GolandProjects/sms-vgateway/logging      ~/cmcc/sms-vgateway/
cp -rf ~/GolandProjects/sms-vgateway/snowflake    ~/cmcc/sms-vgateway/
cp -rf ~/GolandProjects/sms-vgateway/snowflake32  ~/cmcc/sms-vgateway/

