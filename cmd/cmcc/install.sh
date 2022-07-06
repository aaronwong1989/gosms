#!/bin/sh

go clean
go mod tidy
# 编译
go build -trimpath
go test -v server_test.go -test.run TestClient -c

mkdir -p ~/smsvg
mv cmcc ~/smsvg
mv main.test ~/smsvg
cp start.sh ~/smsvg
cp -f cmcc.yaml ~/.cmcc.yaml
cp -rf ../../cmcc ~/smsvg/sms-vgateway/
