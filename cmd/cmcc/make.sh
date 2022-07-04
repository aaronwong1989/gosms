#!/bin/sh

# 在该文件所处当前目录执行（cmd/cmcc 目录）

# 如果你想在Windows 32位系统下运行
# CGO_ENABLED=0 GOOS=windows GOARCH=386 go build
# 如果你想在Windows 64位系统下运行
# CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build
# 如果你想在Linux 32位系统下运行
# CGO_ENABLED=0 GOOS=linux GOARCH=386 go build
# 如果你想在Linux 64位系统下运行
# CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build
# 如果你想在 本机环境 运行
go build

tar -zcf cmcc.tar.gz cmcc cmcc.yaml start.sh
