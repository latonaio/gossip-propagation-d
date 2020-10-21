BINARY=gossip

build:
	GOOS=linux GOARCH=arm go build ./cmd/gossip

install: build
	$(eval defaultuser := $(shell id -u -n))
	sudo cp $(BINARY) /usr/local/bin
	sudo cp ./gossip-propagation-d.service.base ./gossip-propagation-d.service
	sed -i -e "s/USER/$(defaultuser)/g" ./gossip-propagation-d.service
	sudo cp ./gossip-propagation-d.service /etc/systemd/system/
	-sudo mkdir /var/local/distributed-service-discovery
	sudo chown $(defaultuser):$(defaultuser) /var/local/distributed-service-discovery
	sudo systemctl daemon-reload
	sudo systemctl enable gossip-propagation-d.service

delivery: build
	scp $(BINARY) vega:~/
	scp $(BINARY) sirius:~/

get-etcd:
	@echo "---------------------------------------------------------------"
	@etcdctl --endpoints=127.0.0.1:13380 get --prefix "/"
	@echo "---------------------------------------------------------------"
	@etcdctl --endpoints=127.0.0.1:13381 get --prefix "/"
	@echo "---------------------------------------------------------------"
	@etcdctl --endpoints=127.0.0.1:13382 get --prefix "/"
	@echo "---------------------------------------------------------------"

watch-etcd:
	bash ./shell/watch-etcd.sh

etcd-reset:
	# vm-1
	-etcdctl --endpoints=127.0.0.1:13380 del --prefix /
	# vm-2
	-etcdctl --endpoints=127.0.0.1:13381 del --prefix /
	# vm-3
	-etcdctl --endpoints=127.0.0.1:13382 del --prefix /
	# vm-4
	-etcdctl --endpoints=127.0.0.1:13383 del --prefix /

etcd-init: etcd-reset
	# vm-1
	etcdctl --endpoints=127.0.0.1:13380 put /Device/vm-1/0 'vm-1'
	etcdctl --endpoints=127.0.0.1:13380 put /Pod/vm-1/0/ui-frontend-5dcf4958ff-zrkjl 'ui-frontend-5dcf4958ff-zrkjl'
	# vm-2
	etcdctl --endpoints=127.0.0.1:13381 put /Device/vm-2/0 'vm-2'
	etcdctl --endpoints=127.0.0.1:13381 put /Pod/vm-2/0/pull-container-image-to-edge-644b5f574b-fc9qb 'pull-container-image-to-edge-644b5f574b-fc9qb'
	# vm-3
	etcdctl --endpoints=127.0.0.1:13382 put /Device/vm-3/0 'vm-3'
	etcdctl --endpoints=127.0.0.1:13382 put /Pod/vm-3/0/mysql-5f9974cf76-zff2s 'mysql-5f9974cf76-zff2s'
	# vm-4
	etcdctl --endpoints=127.0.0.1:13383 put /Device/vm-4/0 'vm-4'
	etcdctl --endpoints=127.0.0.1:13383 put /Pod/vm-4/0/rust-5f997421-zff2s 'rust-5f997421-zff2s'

# run-vm-2-manual-discovery:
# 	# TODO: change ip dynamically
# 	go run main.go -n vm-2 -j true -i 192.168.10.47:10039 -p 10040 -e 13379 -d

# run-vm-3-manual-discovery:
# 	# TODO: change ip dynamically
# 	go run main.go -n vm-3 -j true -i 192.168.10.47:10039 -p 10041 -e 14379 -d

# run-vm-2-auto-discovery:
# 	# TODO: change ip dynamically
# 	go run main.go -n vm-2 -j true -p 10040 -ep 13379 -d

# run-vm-3-auto-discovery:
# 	# TODO: change ip dynamically
# 	go run main.go -n vm-3 -j true -p 10041 -ep 14379 -d