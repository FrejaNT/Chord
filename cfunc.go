package main

import (
	"fmt"
	"log"
	"math/big"
)

// This file contains most of the standard chord functions and any related RPC methods

func (n *Node) create() {
	n.pNode.n = nil
	n.addSuccNode(n)
}

func (n *Node) join(n_ *Node) {
	n.pNode.n = nil
	succ, err := n.find(n.id, n_)
	if err != nil {
		log.Fatal(err.Error())
	}
	n.addSuccNode(succ)
}

const numberOfFind = 10

func (n *Node) find(id *big.Int, start *Node) (*Node, error) {
	found := false
	nextNode := start
	i := 0
	for !found && i < numberOfFind {
		args := Find_successor_Args{Id: id.Bytes()}
		reply := Find_successor_Reply{}
		if err := n.dial("Node.Find_successor", &args, &reply, nextNode.addr); err != nil {
			n.removeSuccNode(nextNode)
			i++
			continue
		}
		found = reply.OK

		nextNode = &Node{id: convToBig(reply.Id), addr: reply.Addr}
		i++
	}
	if found {
		return nextNode, nil
	}

	return nil, fmt.Errorf("cant find nobody")
}

// find successor (CALL)
func (n *Node) Find_successor(args *Find_successor_Args, reply *Find_successor_Reply) error {
	aid := convToBig(args.Id)
	for i := 0; i < n.r; i++ {
		ns := n.getSuccessor(i)
		if ns == nil {
			continue
		}

		if between(n.id, aid, ns.id, true) {
			reply.Id = ns.id.Bytes()
			reply.Addr = ns.addr
			reply.OK = true
			return nil
		}
	}
	n_, _ := n.closestPrecedingNode(aid)
	reply.Id = n_.id.Bytes()
	reply.Addr = n_.addr
	reply.OK = false

	return nil
}

func (n *Node) closestPrecedingNode(id *big.Int) (*Node, bool) {
	for i := len(n.finger) - 1; i >= 0; i-- {
		if n.finger[i] == nil {
			continue
		}
		if between(n.id, n.finger[i].id, id, true) {
			return n.finger[i], false
		}
	}
	n_ := n.getSuccessor(0)
	return n_, true
}

func (n *Node) stabilize() {
	for i := 0; i < n.r; i++ {
		n.sNodes.mu.Lock()
		s_ := n.sNodes.n[i]
		if s_ == nil {
			n.sNodes.mu.Unlock()
			continue
		}
		sid := new(big.Int) // trying to be safe, getting "deep" copy before unlocking
		*sid = *s_.id
		s := &Node{id: sid, addr: s_.addr}
		n.sNodes.mu.Unlock()
		args := Call_Stabilize_Args{}
		reply := Call_Stabilize_Reply{}
		if err := n.dial("Node.Call_Stabilize", &args, &reply, s.addr); err != nil {
			n.removeSuccNode(s)
			continue
		}

		var new *Node
		args_ := Notify_Args{Id: n.id.Bytes(), Addr: n.addr}
		reply_ := Notify_Reply{}

		if reply.OK && between(n.id, convToBig(reply.Id), s.id, false) {
			new = &Node{id: convToBig(reply.Id), addr: reply.Addr}
			n.addSuccNode(new)
		} else {
			new = s_
		}
		if err := n.dial("Node.Notify", &args_, &reply_, new.addr); err != nil {
			n.removeSuccNode(new)
			continue
		}
		n.addSuccessors(reply_.SIds, reply_.SAddrs) // Notify gets list of successors of new successor
	}
}

func (n *Node) Call_Stabilize(args *Call_Stabilize_Args, reply *Call_Stabilize_Reply) error {
	n.pNode.mu.Lock()
	if n.pNode.n == nil {
		reply.OK = false
		n.pNode.mu.Unlock()
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
	if n.pNode.n == nil || between(n.pNode.n.id, convToBig(args.Id), n.id, false) {
		n.pNode.n = &Node{id: convToBig(args.Id), addr: args.Addr}
		go func() {
			args := Get_Files_Args{Start: n.pNode.n.id.Bytes(), End: n.id.Bytes()}
			reply := Get_Files_Reply{}
			n.dial("Node.GetFiles", &args, &reply, n.getSuccessor(0).addr)
			n.addFiles(reply.Names, reply.Data, nil)
		}()
	}
	n.pNode.mu.Unlock()

	// get lists of successor ids/addresses
	n.sNodes.mu.Lock()
	sIds := make([][]byte, n.r)
	sAddrs := make([]string, n.r)

	for i := 0; i < n.r; i++ {
		if n.sNodes.n[i] == nil {
			continue
		}
		sIds[i] = n.sNodes.n[i].id.Bytes()
		sAddrs[i] = n.sNodes.n[i].addr
	}

	reply.SIds = sIds
	reply.SAddrs = sAddrs
	n.sNodes.mu.Unlock()

	return nil
}

// fix fingers
func (n *Node) fix_fingers() {
	n.index += 1
	if n.index >= M {
		n.index = 1
	}
	fid := jump(n.addr, n.index)
	succ, err := n.find(fid, n)
	if err != nil {
		n.finger[n.index] = nil
		return
	}
	n.finger[n.index] = succ

}

// check predecessor (CALL)
func (n *Node) checkPredecessor() {
	n.pNode.mu.Lock()
	if n.pNode.n == nil {
		n.pNode.mu.Unlock()
		return
	}
	paddr := n.pNode.n.addr
	n.pNode.mu.Unlock()
	args := Call_predecessor_Args{}
	reply := Call_predecessor_Reply{}

	err := n.dial("Node.Call_predecessor", &args, &reply, paddr)

	if err != nil || !reply.OK {
		n.pNode.mu.Lock()
		n.pNode.n = nil
		n.pNode.mu.Unlock()
	}
}

func (n *Node) Call_predecessor(args *Call_predecessor_Args, reply *Call_predecessor_Reply) error {
	reply.OK = true
	return nil
}
