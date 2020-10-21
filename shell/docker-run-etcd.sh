#!/bin/bash

DEVICE=$1
PORT=$2

sh ./docker-stop-etcd.sh ${DEVICE} ${PORT}

docker -H tcp://127.0.0.1:${PORT} run -itd \
  -p 2379:2379 \
  -p 2380:2380 \
  --mount type=bind,source=/tmp/etcd.tmp,destination=/etcd-data \
  --name etcd-gcr-${DEVICE}-v3.4.7 \
  gcr.io/etcd-development/etcd:v3.4.7 \
  /usr/local/bin/etcd \
  --name s1 \
  --data-dir /etcd-data \
  --listen-client-urls http://0.0.0.0:2379 \
  --advertise-client-urls http://0.0.0.0:2379 \
  --listen-peer-urls http://0.0.0.0:2380 \
  --initial-advertise-peer-urls http://0.0.0.0:2380 \
  --initial-cluster s1=http://0.0.0.0:2380 \
  --initial-cluster-token tkn \
  --initial-cluster-state new \
  --log-level info \
  --logger zap \
  --log-outputs stderr