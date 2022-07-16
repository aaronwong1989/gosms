#!/bin/sh

go clean
go mod tidy
# 编译
go build -trimpath -o smgp.ismg
go test -v server_test.go -test.run TestClient -c

mkdir -p ~/smgp
mv smgp.ismg ~/smgp/
mv main.test ~/smgp/
cp start.sh ~/smgp/
cp -f smgp.yaml ~/.smgp.yaml
