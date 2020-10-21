package app

import (
	"fmt"
	"math/rand"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"
	"unsafe"

	"bitbucket.org/latonaio/gossip-propagation-d/internal/packet"
	"bitbucket.org/latonaio/gossip-propagation-d/pkg/log"
	"bitbucket.org/latonaio/gossip-propagation-d/pkg/my_etcd"
	"github.com/hashicorp/memberlist"
	"github.com/spf13/cobra"
)

type Server struct {
	flags    *GossipPropagationFlags
	interval *time.Ticker
	db       *packet.Etcd
	cache    *packet.Cache
	mu       sync.Mutex
}

// NewGossipPropagationCommand creates a *cobra.Command object with default parameters
func NewGossipPropagationCommand() *cobra.Command {
	gossipPropagationFlags := newGossipPropagationFlags()

	cmd := &cobra.Command{
		Use:   componentGossipPropagation,
		Short: "run gossip propagation",
		// entrypoint
		Run: func(cmd *cobra.Command, args []string) {
			f := gossipPropagationFlags
			log.SetFormat(packageName, f.DebugMode)

			if err := my_etcd.GetInstance().CreatePool(f.EtcdHost, f.EtcdPort); err != nil {
				log.Fatalf("can't connect to etcd: %v", err)
			}

			cluster := NewCluster(f)
			go cluster.Join()

			server := NewServer(f)
			server.start(cluster)
		},
	}

	fs := cmd.Flags()
	gossipPropagationFlags.set(fs)

	return cmd
}

func NewServer(f *GossipPropagationFlags) *Server {
	return &Server{
		flags:    f,
		interval: time.NewTicker(time.Duration(f.ConsystencyInterval) * time.Second),
		db:       packet.GetEtcdInstance(),
		cache:    packet.GetCacheInstance(),
	}
}

func (s *Server) start(c *Cluster) error {
	signalCh := make(chan os.Signal)
	signal.Notify(signalCh, syscall.SIGTERM)

	fowardCh := make(chan *packet.Cache, 1)
	go func() {
		for {
			select {
			case fowardCache := <-fowardCh:
				fowardCache.AddSharedNodes([]*memberlist.Node{c.memberlist.LocalNode()})
				fowardCache.OverrideFromNode(c.memberlist.LocalNode())
				fowardCache.OverrideSendTypeToFoward()

				if err := gossip(c, fowardCache, s.flags.GossipNodeNum); err != nil {
					log.Errorf("failed to foward packets. %v", err)
				}
			}
		}
	}()

loop:
	for {
		select {
		// update db when recieve remote cache
		case cacheBytes := <-c.delegate.cacheCh:
			remoteCache := packet.UnmarshalCache(cacheBytes)
			if err := s.db.MergeKVCacheList(&remoteCache.KvCacheList); err != nil {
				log.Errorf("failed to merge kv cache list on etcd. %v", err)
			}
			log.Infof("recieved packets <from: %s type: %s>", remoteCache.FromNode.Name, packet.GetSendTypeName(remoteCache.SendType))
			log.Debugf("recieved contents <keyList: %v>", remoteCache.KvCacheKeyList)

			fowardCh <- remoteCache

		// cluster lifecycle
		case joinNode := <-c.events.nodeJoin:
			if joinNode.Name != c.flags.MyDevice {
				c.isJoined = true
				if err := s.sendBulkPackets(c); err != nil {
					log.Errorf("failed to send bulk packets. %v", err)
				}
			} else {
				s.disableExceptMyRecords(joinNode)
			}
		case leaveNode := <-c.events.nodeLeave:
			if err := s.db.DisableRecordByDevice(leaveNode); err != nil {
				log.Errorf("failed to disable record <node: %s> %v", leaveNode.Name, err)
			}

		// interval
		case <-s.interval.C:
			if err := s.sendBulkPackets(c); err != nil {
				log.Errorf("failed to send bulk packets. %v", err)
			}

		// event driven
		case watchResponse := <-s.db.Watch():
			for _, event := range watchResponse.Events {
				key := *(*string)(unsafe.Pointer(&event.Kv.Key))
				if event.Type.String() == "PUT" { // PUT event
					if err := s.sendPacket(c, key); err != nil {
						log.Errorf("failed to send a packet. %v", err)
					}
				}
			}

		// os signal
		case signal := <-signalCh:
			log.Infof("recieved signal: %s", signal.String())
			break loop
		}
	}

	my_etcd.GetInstance().Close()
	return nil
}

func (s *Server) sendBulkPackets(c *Cluster) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if err := s.cache.UpdateWithMyRecords(s.db, c.memberlist.LocalNode()); err != nil {
		return fmt.Errorf("%v", err)
	}

	if err := gossip(c, s.cache, s.flags.GossipNodeNum); err != nil {
		return fmt.Errorf("%v", err)
	}

	return nil
}

func (s *Server) sendPacket(c *Cluster, key string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if err := s.cache.UpdateWithARecord(key, c.memberlist.LocalNode()); err != nil {
		return fmt.Errorf("%v", err)
	}

	if err := gossip(c, s.cache, s.flags.GossipNodeNum); err != nil {
		return fmt.Errorf("%v", err)
	}

	return nil
}

func (s *Server) disableExceptMyRecords(myNode *memberlist.Node) error {
	recordListExceptMine, err := s.db.FetchRecordList(myNode, false)
	if err != nil {
		return fmt.Errorf("failed to fetch records from etcd %v", err)
	}

	for _, record := range recordListExceptMine {
		beforeRecordKey := record.Key
		afterRecordKey := strings.Replace(beforeRecordKey, packet.GetRunningStatusIndex(), packet.GetStoppedStatusIndex(), -1)

		if err := s.db.DisableRecord(beforeRecordKey, afterRecordKey, record.Value); err != nil {
			log.Errorf("failed to disable record. %v", err)
			continue
		}
		log.Infof("disabled status <key: %s -> %s>", beforeRecordKey, afterRecordKey)
	}

	return nil
}

func gossip(c *Cluster, cache *packet.Cache, gossipNodeNum int) error {
	gossipNodes := kRandomNodes(gossipNodeNum, c.GetAvailableNodes(cache.SharedNodes))
	if len(gossipNodes) == 0 {
		log.Debugf("cancel to send packets. every node recieved packat")
		return nil
	}

	for _, node := range gossipNodes {
		if err := c.memberlist.SendReliable(node, cache.Marshal()); err != nil {
			return fmt.Errorf("failed to send packet <to: %s> %v", node.Name, err)
		}
		log.Infof("sent packets <to: %s, type:%s>", node.Name, packet.GetSendTypeName(cache.SendType))
		log.Debugf("sent contents <keyList: %v>", cache.KvCacheKeyList)
	}
	return nil
}

func kRandomNodes(k int, availableNodes []*memberlist.Node) []*memberlist.Node {
	n := len(availableNodes)
	kNodes := make([]*memberlist.Node, 0, k)

	for i := 0; i < n && len(kNodes) < k; i++ {
		idx := randomOffset(n)
		node := availableNodes[idx]

		// Check if we have this node already
		isNode := false
		for j := 0; j < len(kNodes); j++ {
			if node == kNodes[j] {
				isNode = true
				break
			}
		}

		if !isNode {
			kNodes = append(kNodes, node)
		}
	}

	return kNodes
}

// Returns a random offset between 0 and n
func randomOffset(n int) int {
	if n == 0 {
		return 0
	}
	return int(rand.Uint32() % uint32(n))
}
