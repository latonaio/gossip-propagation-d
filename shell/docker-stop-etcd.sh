#!/bin/bash

DEVICE=$1
PORT=$2

docker -H tcp://127.0.0.1:${PORT} ps -a | grep etcd-gcr-${DEVICE}-v3.4.7 > /dev/null
ret=`echo $?`
if [ $ret = 0 ]; then
    docker -H tcp://127.0.0.1:${PORT} stop etcd-gcr-${DEVICE}-v3.4.7 > /dev/null && \
    docker -H tcp://127.0.0.1:${PORT} rm etcd-gcr-${DEVICE}-v3.4.7 > /dev/null
fi