#!/bin/sh

go clean
go mod tidy

# 如果你想在Windows 32位系统下运行
# CGO_ENABLED=0 GOOS=windows GOARCH=386 go build -trimpath -o cmpp.ismg

# 如果你想在Windows 64位系统下运行
# CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -trimpath -o cmpp.ismg

# 如果你想在Linux 32位系统下运行
# CGO_ENABLED=0 GOOS=linux GOARCH=386 go build -trimpath -o cmpp.ismg

# 如果你想在Linux 64位系统下运行
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -trimpath -o cmpp.ismg

# 如果你想在Linux arm64系统下运行
# CGO_ENABLED=0 GOOS=linux GOARM=7 GOARCH=arm64 go build -trimpath -o cmpp.ismg

# 如果你想在 本机环境 运行
# go build -trimpath -o cmpp.ismg

# 制作软件发布包
chmod +x cmpp.ismg
chmod +x start.sh
tar -zcvf cmpp.ismg.tar.gz cmpp.ismg cmpp.yaml start.sh
