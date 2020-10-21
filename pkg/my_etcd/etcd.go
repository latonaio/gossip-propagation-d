package my_etcd

import (
	"context"
	"strconv"
	"time"

	"go.etcd.io/etcd/clientv3"
	"go.etcd.io/etcd/etcdserver/api/v3rpc/rpctypes"
)

var instance = &Etcd{}

type Etcd struct {
	address []string
	client  *clientv3.Client
}

const (
	address        = "localhost"
	port           = 2379
	dialTimeout    = 2 * time.Second
	requestTimeout = 10 * time.Second
)

func GetInstance() *Etcd {
	return instance
}

func (e *Etcd) CreatePool(address string, port int) error {
	client, err := clientv3.New(clientv3.Config{
		Endpoints:   []string{address + ":" + strconv.Itoa(port)},
		DialTimeout: dialTimeout,
	})
	if err != nil {
		return err
	}
	e.client = client

	return nil
}

func (e *Etcd) Close() error {
	err := e.client.Close()
	if err != nil {
		return err
	}
	return nil
}

func (e *Etcd) Get(key string) (*KV, error) {
	ctx, _ := context.WithTimeout(context.Background(), requestTimeout)

	resp, err := e.client.Get(ctx, key)
	if err != nil {
		switch err {
		case context.Canceled:
			return nil, context.Canceled
		case context.DeadlineExceeded:
			return nil, context.DeadlineExceeded
		case rpctypes.ErrEmptyKey:
			return nil, rpctypes.ErrEmptyKey

		default:
			return nil, err
		}
	}

	return newKV(resp.Kvs[0].Key, resp.Kvs[0].Value), nil
}

func (e *Etcd) GetWithPrefix(key string) ([]KV, error) {
	ctx, _ := context.WithTimeout(context.Background(), requestTimeout)

	resp, err := e.client.Get(ctx, key, clientv3.WithPrefix())
	if err != nil {
		switch err {
		case context.Canceled:
			return nil, context.Canceled
		case context.DeadlineExceeded:
			return nil, context.DeadlineExceeded
		case rpctypes.ErrEmptyKey:
			return nil, rpctypes.ErrEmptyKey

		default:
			return nil, err
		}
	}

	var kvList []KV
	for _, ev := range resp.Kvs {
		kv := newKV(ev.Key, ev.Value)
		kvList = append(kvList, *kv)
	}
	return kvList, nil
}

func (e *Etcd) Put(key string, value string) error {
	ctx, _ := context.WithTimeout(context.Background(), requestTimeout)

	if _, err := e.client.Put(ctx, key, value); err != nil {
		switch err {
		case context.Canceled:
			return context.Canceled
		case context.DeadlineExceeded:
			return context.DeadlineExceeded
		case rpctypes.ErrEmptyKey:
			return rpctypes.ErrEmptyKey

		default:
			return err
		}
	}

	return nil
}

func (e *Etcd) Delete(key string) error {
	ctx, _ := context.WithTimeout(context.Background(), requestTimeout)

	if _, err := e.client.Delete(ctx, key); err != nil {
		return err
	}

	return nil
}

func (e *Etcd) WatchWithPrefix(key string) clientv3.WatchChan {
	return e.client.Watch(context.Background(), key, clientv3.WithPrefix())
}

type KV struct {
	Key   string
	Value string
}

func newKV(key []byte, value []byte) *KV {
	return &KV{
		Key:   string(key),
		Value: string(value),
	}
}
