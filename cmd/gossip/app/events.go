package app

import (
	"bitbucket.org/latonaio/gossip-propagation-d/pkg/log"
	"github.com/hashicorp/memberlist"
)

type Events struct {
	nodeJoin   chan *memberlist.Node
	nodeUpdate chan *memberlist.Node
	nodeLeave  chan *memberlist.Node
}

func NewEvents() *Events {
	return &Events{
		nodeJoin:   make(chan *memberlist.Node, 1),
		nodeLeave:  make(chan *memberlist.Node, 1),
		nodeUpdate: make(chan *memberlist.Node, 1),
	}
}

func (e *Events) NotifyJoin(node *memberlist.Node) {
	log.Debugf("notify join %s", log.GetFromAddress(node.Name, node.Addr.To4().String()))
	e.nodeJoin <- node
}

func (e *Events) NotifyLeave(node *memberlist.Node) {
	log.Debugf("notify leave %s", log.GetFromAddress(node.Name, node.Addr.To4().String()))
	e.nodeLeave <- node
}

func (e *Events) NotifyUpdate(node *memberlist.Node) {
	log.Debugf("notify update %s", log.GetFromAddress(node.Name, node.Addr.To4().String()))
	e.nodeUpdate <- node
}
