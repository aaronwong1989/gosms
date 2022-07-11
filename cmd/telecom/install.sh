#!/bin/sh

go clean
go mod tidy
# 编译
go build -trimpath
go test -v server_test.go -test.run TestClient -c

mkdir -p               ~/telecom
mv telecom             ~/telecom/
mv main.test           ~/telecom/
cp start.sh            ~/telecom/
cp -f telecom.yaml     ~/.telecom.yaml

# 拷贝源码
cp -rf ~/GolandProjects/sms-vgateway/cmcc         ~/telecom/sms-vgateway/
cp -rf ~/GolandProjects/sms-vgateway/telecom      ~/telecom/sms-vgateway/
cp -rf ~/GolandProjects/sms-vgateway/comm         ~/telecom/sms-vgateway/


