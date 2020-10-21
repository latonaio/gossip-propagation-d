package app

import (
	"fmt"
	"io/ioutil"
	"strings"
	"time"

	"bitbucket.org/latonaio/gossip-propagation-d/pkg/log"
	"github.com/avast/retry-go"
	"github.com/hashicorp/memberlist"
)

type Cluster struct {
	flags      *GossipPropagationFlags
	events     *Events
	delegate   *Delegate
	isJoined   bool
	memberlist *memberlist.Memberlist
}

const (
	retryInterval  = 10
	joinRetryCount = 30
)

func NewCluster(f *GossipPropagationFlags) *Cluster {
	events := NewEvents()
	delegate := NewDelegate()

	memberlist, err := newMemberlist(f, delegate, events)
	if err != nil {
		log.Fatalf("[FATAL] %v", err)
	}

	return &Cluster{
		flags:      f,
		events:     events,
		delegate:   delegate,
		isJoined:   false,
		memberlist: memberlist,
	}
}

func newMemberlist(f *GossipPropagationFlags, delegate *Delegate, events *Events) (*memberlist.Memberlist, error) {
	conf := memberlist.DefaultLocalConfig()

	conf.GossipInterval = 500 * time.Millisecond
	conf.Name = f.MyDevice
	conf.BindPort = f.Port
	conf.AdvertisePort = conf.BindPort

	if f.MyIP != "" {
		conf.BindAddr = f.MyIP
	}

	conf.Delegate = delegate
	conf.Events = events

	log.SetFormat("memberlist", f.DebugMode)
	conf.Logger = log.Get()

	return memberlist.Create(conf)
}

func (c *Cluster) Join() error {
	if !c.flags.Join {
		log.Debug("no need start join loop")
		return nil
	}

	specifiedIP := c.flags.IP
	if len(specifiedIP) > 0 {
		if _, err := c.memberlist.Join([]string{strings.TrimRight(specifiedIP, "\n")}); err != nil {
			log.Fatalf("can't recieve join response from=<specifiedIP: %v>", specifiedIP)
		}
		return nil
	}

	if err := retry.Do(
		func() error {
			if c.isJoined {
				log.Info("cluster has created. stop to retry request join")
				return nil
			}

			files, err := ioutil.ReadDir(c.flags.MonitorDir)
			if err != nil {
				return fmt.Errorf("can't find IP file list")
			}

			for _, file := range files {
				if !file.IsDir() {
					addr := strings.Replace(file.Name(), ".txt", "", 1)
					if _, err := c.memberlist.Join([]string{strings.TrimRight(addr, "\n")}); err != nil {
						log.Errorf("%v", err)
						continue
					}
					return nil
				}
			}
			return fmt.Errorf("can't find cluster member")
		},
		retry.DelayType(func(n uint, config *retry.Config) time.Duration {
			log.Infof("retry to join cluster after %d seconds", retryInterval)
			return retryInterval * time.Second
		}),
		retry.Attempts(joinRetryCount),
	); err != nil {
		log.Info("timeout join loop. start standalone mode...")
		return nil
	}

	return nil
}

func (c *Cluster) GetAvailableNodes(sharedNodes []*memberlist.Node) []*memberlist.Node {
	var availableNodes []*memberlist.Node
	for _, node := range c.memberlist.Members() {
		isNode := false
		for _, sharedNode := range sharedNodes {
			if node.Name == sharedNode.Name {
				isNode = true
				break
			}
		}
		if !isNode {
			availableNodes = append(availableNodes, node)
		}
	}
	return availableNodes
}
