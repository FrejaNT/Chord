package main

func (n *Node) Create() {
	n.data.pNode = nil
	n.data.sNodes = append(n.data.sNodes, &Node{id: n.id, addr: n.addr})
}

func (n *Node) Join(n_ *Node) {
	n.data.pNode = nil
	args := Find_successor_Args{Id: n.id}
	reply := Find_successor_Reply{}
	n.Dial("Node.Find_successor", args, reply, n_.addr) // error
	n.AddSuccNode(&Node{id: reply.Id, addr: reply.Addr})
}

// find successor (CALL)
func (n *Node) Find_successor(args *Find_successor_Args, reply *Find_successor_Reply) error {
	for _, s := range n.data.sNodes {
		if args.Id <= s.id {
			reply.Id = s.id
			reply.Addr = s.addr
			return nil
		}
	}
	n_, this := n.Closest_preceding_node(args.Id)
	if this {
		reply.Id = n.id
		reply.Addr = n.addr
		return nil
	}
	n.Dial("Node.Find_Successor", args, reply, n_.addr)
	return nil
}

// closest preceding node
func (n *Node) Closest_preceding_node(id uint64) (*Node, bool) {
	for i := len(n.data.finger) - 1; i >= 0; i-- {
		if n.data.finger[i].n.id > n.id && n.data.finger[i].n.id < id { // n.finger[i].n.id <= id ???
			return n.data.finger[i].n, false
		}
	}
	return n, true
}

// stabilize (CALL)
func (n *Node) Stabilize() {
	args := Call_Stabilize_Args{}
	reply := Call_Stabilize_Reply{}
	n.Dial("Node.Call_Stabilize", &args, &reply, n.data.sNodes[0].addr) // error
	if reply.Id > n.id && reply.Id < n.data.sNodes[0].id {
		n.AddSuccNode(&Node{id: reply.Id, addr: reply.Addr})
	}
	args_ := Notify_Args{Id: n.id, Addr: n.addr}
	reply_ := Notify_Reply{}
	n.Dial("Node.Notify", &args_, &reply_, n.data.sNodes[0].addr)
}

func (n *Node) Call_Stabilize(args *Call_Stabilize_Args, reply *Call_Stabilize_Reply) {
	reply.Id = n.data.pNode.id
	reply.Addr = n.data.pNode.addr
}

// notify (CALL)
func (n *Node) Notify(args *Notify_Args, reply *Notify_Reply) {
	if n.data.pNode == nil || args.Id > n.data.pNode.id && args.Id < n.id {
		n.data.pNode = &Node{id: args.Id, addr: args.Addr}
	}
}

// fix fingers
func (n *Node) Fix_fingers() {
	for i := range n.data.finger {
		args := Find_successor_Args{Id: n.data.finger[i].i}
		reply := Find_successor_Reply{}
		n.Find_successor(&args, &reply) // error
	}
}

// check predecessor (CALL)
func (n *Node) Check_predecessor() {
	args := Call_predecessor_Args{}
	reply := Call_predecessor_Reply{}
	n.Dial("Node.Call_predecessor", args, reply, n.data.pNode.addr)
	if reply.OK {
		return
	}
	n.data.pNode = nil
}
func (n *Node) Call_predecessor(args *Call_predecessor_Args, reply *Call_predecessor_Reply) {
	reply.OK = true
}
