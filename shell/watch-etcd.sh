#!/bin/bash

while :
do
    echo ""
    date "+%Y/%m/%d %H:%M:%S"
    echo "---------------------------------------------"
    echo "[vm-1]"
    etcdctl --endpoints=127.0.0.1:13380 get --prefix "/"
    echo "---------------------------------------------"
    echo "[vm-2]"
    etcdctl --endpoints=127.0.0.1:13381 get --prefix "/"
    echo "---------------------------------------------"
    echo "[vm-3]"
    etcdctl --endpoints=127.0.0.1:13382 get --prefix "/"
    echo "---------------------------------------------"
    echo ""
    sleep 5
done