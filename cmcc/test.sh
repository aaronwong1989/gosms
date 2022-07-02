#!/bin/zsh

echo "" >test.result.txt

i=$1

while ((i > 0)); do
  go test server_test.go &
  ((i--))
done

tail -f test.result.txt

