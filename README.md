# What is this?

複数端末間でデータ同期するサービスです

## dependencies

- distributed-service-discovery

## how to setup
```
$ git clone git@bitbucket.org:latonaio/gossip-propagation-d.git -b v0.9.2 && cd gossip-propagation-d
$ make install
```

## how to run
### via systemd
```
$ sudo systemctl start gossip-propagation-d.service
```

### manually
```
$ gossip -j
```