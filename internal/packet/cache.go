package packet

import (
	"encoding/json"
	"fmt"

	"bitbucket.org/latonaio/gossip-propagation-d/pkg/my_etcd"
	"github.com/hashicorp/memberlist"
)

const (
	origin = iota
	foward
	unknown
)

var sharedCacheInstance *Cache = &Cache{}

type Cache struct {
	FromNode       *memberlist.Node   `json:"fromNode"`
	SharedNodes    []*memberlist.Node `json:"sharedNodes"`
	KvCacheList    []KVCache          `json:"kvCacheList"`
	KvCacheKeyList []string           `json:"kvCacheKeyList"`
	SendType       int                `json:"sendType"`
}

type KVCache struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

func GetCacheInstance() *Cache {
	return sharedCacheInstance
}

func UnmarshalCache(bytes []byte) *Cache {
	cache := new(Cache)
	if err := json.Unmarshal(bytes, &cache); err != nil {
		return nil
	}
	return cache
}

func GetSendTypeName(sendType int) string {
	switch sendType {
	case origin:
		return "origin"
	case foward:
		return "foward"
	}
	return "unknown"
}

func (c *Cache) Clear() {
	c.FromNode = nil
	c.SharedNodes = make([]*memberlist.Node, 0)
	c.KvCacheList = make([]KVCache, 0)
	c.KvCacheKeyList = make([]string, 0)
	c.SendType = unknown
}

func (c *Cache) UpdateWithARecord(key string, myNode *memberlist.Node) error {
	c.Clear()

	record, err := my_etcd.GetInstance().Get(key)
	if err != nil {
		return fmt.Errorf("failed to fetch a record from etcd <key: %s> %v", key, err)
	}

	c.insert(record.Key, record.Value)
	c.FromNode = myNode
	c.SendType = origin
	c.AddSharedNodes([]*memberlist.Node{myNode})

	return nil
}

func (c *Cache) UpdateWithMyRecords(db *Etcd, myNode *memberlist.Node) error {
	c.Clear()

	myRecordList, err := db.FetchRecordList(myNode, true)
	if err != nil {
		return fmt.Errorf("failed to fetch my records from etcd %v", err)
	}

	for _, record := range myRecordList {
		c.insert(record.Key, record.Value)
	}
	c.FromNode = myNode
	c.SendType = origin
	c.AddSharedNodes([]*memberlist.Node{myNode})

	return nil
}

func (c *Cache) AddSharedNodes(nodes []*memberlist.Node) error {
	for _, node := range nodes {
		// Check if we have this node already
		isNode := false
		for _, member := range c.SharedNodes {
			if node == member {
				isNode = true
			}
		}

		if !isNode {
			c.SharedNodes = append(c.SharedNodes, node)
		}
	}
	return nil
}

func (c *Cache) OverrideFromNode(myNode *memberlist.Node) {
	c.FromNode = myNode
}

func (c *Cache) OverrideSendTypeToFoward() {
	c.SendType = foward
}

func (c *Cache) Marshal() []byte {
	bytes, err := json.Marshal(c)
	if err != nil {
		return []byte("")
	}
	return bytes
}

func (c *Cache) insert(key string, value string) {
	if !c.in(key) {
		c.KvCacheList = append(c.KvCacheList, KVCache{
			Key:   key,
			Value: value,
		})
		c.KvCacheKeyList = append(c.KvCacheKeyList, key)
	}
}

func (c *Cache) in(key string) bool {
	if len(c.KvCacheList) > 0 {
		for _, value := range c.KvCacheList {
			if key == value.Key {
				return true
			}
		}
	}
	return false
}
