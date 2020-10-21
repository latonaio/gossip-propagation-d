package packet

import (
	"fmt"
	"strings"

	"bitbucket.org/latonaio/gossip-propagation-d/pkg/log"
	"bitbucket.org/latonaio/gossip-propagation-d/pkg/my_etcd"
	"github.com/hashicorp/memberlist"
	"go.etcd.io/etcd/clientv3"
)

var sharedEtcdInstance *Etcd = &Etcd{}

type Etcd struct {
	recordList []Record
}

type Record struct {
	Key   string
	Value string
}

const (
	separator          = "/"
	deviceKeyIndex     = 2
	statusKeyIndex     = 3
	runningStatusIndex = "/0"
	stoppedStatusIndex = "/1"
)

func GetEtcdInstance() *Etcd {
	return sharedEtcdInstance
}

func GetRunningStatusIndex() string {
	return runningStatusIndex
}

func GetStoppedStatusIndex() string {
	return stoppedStatusIndex
}

// TODO: should change name of 2nd arg. "only" is a not appropriate name.
func (e *Etcd) FetchRecordList(device *memberlist.Node, only bool) ([]Record, error) {
	var recordsOfDevice []Record
	records, err := my_etcd.GetInstance().GetWithPrefix("/")
	if err != nil {
		return nil, err
	}

	for _, record := range records {
		recordDeviceName := strings.Split(record.Key, separator)[deviceKeyIndex]
		switch only {
		case true:
			if recordDeviceName == device.Name {
				recordsOfDevice = append(
					recordsOfDevice,
					Record{
						Key:   record.Key,
						Value: record.Value,
					},
				)
			}
		case false:
			if recordDeviceName != device.Name {
				recordsOfDevice = append(
					recordsOfDevice,
					Record{
						Key:   record.Key,
						Value: record.Value,
					},
				)
			}
		}
	}

	return recordsOfDevice, nil
}

func (e *Etcd) MergeKVCacheList(kvCacheList *[]KVCache) error {
	records, err := my_etcd.GetInstance().GetWithPrefix("/")
	if err != nil {
		return err
	}

	for _, kvCache := range *kvCacheList {
		// check & delete old status record
		for _, record := range records {
			if excludeKeyStatus(kvCache.Key) == excludeKeyStatus(record.Key) {
				if err := my_etcd.GetInstance().Delete(record.Key); err != nil {
					log.Errorf("can't delete record <key: %s> %v", record.Key, err)
				}
				log.Debugf("deleted old record <key: %s>", record.Key)
				break
			}
		}

		if err := my_etcd.GetInstance().Put(kvCache.Key, kvCache.Value); err != nil {
			return fmt.Errorf("can't put etcd with <key: %s> %v", kvCache.Key, err)
		}
	}
	return nil
}

func (e *Etcd) DisableRecordByDevice(device *memberlist.Node) error {
	records, err := e.FetchRecordList(device, true)
	if err != nil {
		return fmt.Errorf("failed to fetch records from etcd %v", err)
	}

	for _, record := range records {
		beforeRecordKey := record.Key
		afterRecordKey := strings.Replace(beforeRecordKey, runningStatusIndex, stoppedStatusIndex, -1)

		if err := e.DisableRecord(beforeRecordKey, afterRecordKey, record.Value); err != nil {
			log.Errorf("failed to disable record. %v", err)
			continue
		}
		log.Infof("disabled status <node: %s key: %s -> %s>", device.Name, beforeRecordKey, afterRecordKey)
	}

	return nil
}

func (e *Etcd) DisableRecord(beforeRecordKey string, afterRecordKey string, recordValue string) error {
	if err := my_etcd.GetInstance().Delete(beforeRecordKey); err != nil {
		return fmt.Errorf("can't delete record <key: %s> %v", beforeRecordKey, err)
	}

	// override key status 0 -> 1
	if err := my_etcd.GetInstance().Put(afterRecordKey, recordValue); err != nil {
		return fmt.Errorf("can't put etcd with <key: %s> %v", afterRecordKey, err)
	}
	return nil
}

func (e *Etcd) Watch() clientv3.WatchChan {
	return my_etcd.GetInstance().WatchWithPrefix("/")
}

func excludeKeyStatus(key string) string {
	keyStatus := "/" + strings.Split(key, "/")[statusKeyIndex]
	return strings.Replace(key, keyStatus, "", -1)
}
