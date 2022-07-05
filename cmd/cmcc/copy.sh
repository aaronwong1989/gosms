#!/bin/sh

# 设置环境信息

# 编译
go build -trimpath

mv cmcc ~/smsvg/
cp -f cmcc.yaml ~/.cmcc.yaml
