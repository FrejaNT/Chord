package main

import (
	"math/big"
)

func (n *Node) Create() {
	n.pNode.n = nil
	n.AddSuccNode(n)
}

func (n *Node) Join(n_ *Node) {
	n.pNode.n = nil
	args := Find_successor_Args{Id: n.id.Bytes()}
	reply := Find_successor_Reply{}
	success := n.Dial("Node.Find_successor", &args, &reply, n_.addr) // error
	if !success {
		return
	}
	n.AddSuccNode(&Node{id: ConvToBig(reply.Id), addr: reply.Addr})
}

// find successor (CALL)
func (n *Node) Find_successor(args *Find_successor_Args, reply *Find_successor_Reply) error {
	aid := ConvToBig(args.Id)
	n_, this := n.Closest_preceding_node(aid)
	if this {
		reply.Id = n.id.Bytes()
		reply.Addr = n.addr
		return nil
	}
	for _, s := range n.data.sNodes {
		if s == nil {
			continue
		}
		if Between(n.id, aid, s.id, true) {
			reply.Id = s.id.Bytes()
			reply.Addr = s.addr
			success := n.Dial("Node.Find_Successor", &args, &reply, n_.addr)
			if success {
				return nil
			}
		}
	}

	return nil
}

// closest preceding node
func (n *Node) Closest_preceding_node(id *big.Int) (*Node, bool) {
	for i := len(n.ft.finger) - 1; i >= 0; i-- {
		if n.ft.finger[i].n == nil {
			continue
		}
		if Between(n.id, n.ft.finger[i].n.id, id, false) { // n.finger[i].n.id <= id ???
			return n.ft.finger[i].n, false
		}
	}
	return n, true
}

// stabilize (CALL)
func (n *Node) Stabilize() {
	args := Call_Stabilize_Args{}
	reply := Call_Stabilize_Reply{}

	s := n.GetSuccNode()
	if s == nil {
		return
	}
	success := n.Dial("Node.Call_Stabilize", &args, &reply, s.addr)

	if !success {
		n.RemoveSuccNode(s)
		return
	}

	args_ := Notify_Args{Id: n.id.Bytes(), Addr: n.addr}
	reply_ := Notify_Reply{}

	if reply.OK && Between(n.id, ConvToBig(reply.Id), s.id, false) {
		new := &Node{id: ConvToBig(reply.Id), addr: reply.Addr}
		n.AddSuccNode(new)
		n.Dial("Node.Notify", &args_, &reply_, new.addr)
		return
	}
	n.Dial("Node.Notify", &args_, &reply_, s.addr)
}

func (n *Node) Call_Stabilize(args *Call_Stabilize_Args, reply *Call_Stabilize_Reply) error {
	n.pNode.mu.Lock()
	if n.pNode.n == nil {
		reply.OK = false
		return nil
	}
	reply.Id = n.pNode.n.id.Bytes()
	reply.Addr = n.pNode.n.addr
	reply.OK = true
	n.pNode.mu.Unlock()
	return nil
}

// notify (CALL)
func (n *Node) Notify(args *Notify_Args, reply *Notify_Reply) error {
	n.pNode.mu.Lock()
	if n.pNode.n == nil || Between(n.pNode.n.id, ConvToBig(args.Id), n.id, false) {
		n.pNode.n = &Node{id: ConvToBig(args.Id), addr: args.Addr}
	}
	n.pNode.mu.Unlock()
	return nil
}

// fix fingers
func (n *Node) Fix_fingers() {
	for i := range n.ft.finger {
		go func() {
			args := Find_successor_Args{Id: n.ft.finger[i].i.Bytes()}
			reply := Find_successor_Reply{}
			n.Find_successor(&args, &reply)
		}() // error
	}

}

// check predecessor (CALL)
func (n *Node) Check_predecessor() {
	n.pNode.mu.Lock()
	if n.pNode.n == nil {
		return
	}
	paddr := n.pNode.n.addr
	n.pNode.mu.Unlock()
	args := Call_predecessor_Args{}
	reply := Call_predecessor_Reply{}

	success := n.Dial("Node.Call_predecessor", &args, &reply, paddr)
	n.pNode.mu.Lock()
	if !success {
		n.pNode.n = nil
	} // error
	if reply.OK {
		return
	}
	n.pNode.n = nil
	n.pNode.mu.Unlock()
}
func (n *Node) Call_predecessor(args *Call_predecessor_Args, reply *Call_predecessor_Reply) error {
	reply.OK = true
	return nil
}
