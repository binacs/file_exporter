#!/bin/sh

pid_exporter=$(ps -ef | grep exporter | awk '{print $2}')
for i in $pid_exporter; do
        echo "kill exporter $i"
        kill -9 $i
done
