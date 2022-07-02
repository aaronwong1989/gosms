#!/bin/zsh

echo "" >test.result.txt

i=1
v=$#

if ((v == 1)); then
  i=$1
fi
if ((i > 5)); then
  i=5
fi

while ((i > 0)); do
  go test server_test.go &
  ((i--))
done

tail -f test.result.txt
