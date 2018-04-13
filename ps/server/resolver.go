package server

import (
	"fmt"
	"sync"
	"sync/atomic"

	"github.com/tiglabs/baud/proto/metapb"
	"github.com/tiglabs/raft"
)

type nodeRef struct {
	metapb.Node
	refCount int32
}

// NodeResolver resolve NodeID to net.Addr addresses
type NodeResolver struct {
	nodes *sync.Map
}

// NewNodeResolver create NodeResolver
func NewNodeResolver() *NodeResolver {
	return &NodeResolver{
		nodes: new(sync.Map),
	}
}

func (r *NodeResolver) addNode(node *metapb.Node) {
	if node == nil {
		return
	}

	ref := new(nodeRef)
	ref.Node = *node
	obj, _ := r.nodes.LoadOrStore(node.ID, ref)
	atomic.AddInt32(&(obj.(*nodeRef).refCount), 1)
}

func (r *NodeResolver) deleteNode(id metapb.NodeID) {
	if obj, _ := r.nodes.Load(id); obj != nil {
		if count := atomic.AddInt32(&(obj.(*nodeRef).refCount), -1); count <= 0 {
			r.nodes.Delete(id)
		}
	}
}

func (r *NodeResolver) getNode(id metapb.NodeID) (*metapb.Node, error) {
	if obj, _ := r.nodes.Load(id); obj != nil {
		return &(obj.(*nodeRef).Node), nil
	}
	return nil, fmt.Errorf("Cannot get node network information, nodeID=[%d]", id)
}

// NodeAddress resolve NodeID to net.Addr addresses.
func (r *NodeResolver) NodeAddress(nodeID uint64, stype raft.SocketType) (string, error) {
	node, err := r.getNode(metapb.NodeID(nodeID))
	if err != nil {
		return "", err
	}

	switch stype {
	case raft.HeartBeat:
		return node.RaftAddrs.HeartbeatAddr, nil
	case raft.Replicate:
		return node.RaftAddrs.ReplicateAddr, nil
	default:
		return "", fmt.Errorf("Unknown socket type[%v]", stype)
	}
}