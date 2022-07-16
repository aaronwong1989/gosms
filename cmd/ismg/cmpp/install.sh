#!/bin/sh

go clean
go mod tidy
# 编译
go build -trimpath -o cmpp.ismg
go test -v server_test.go -test.run TestClient -c

mkdir -p ~/cmpp
mv cmpp.ismg ~/cmpp/
mv main.test ~/cmpp/
cp start.sh ~/cmpp/
cp -f cmpp.yaml ~/.cmpp.yaml
