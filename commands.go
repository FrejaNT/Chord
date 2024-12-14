package main

import (
	"fmt"
)

func (n *Node) PrintState() {
	fmt.Println()
	fmt.Println("Node:")
	n.printNode()
	fmt.Println()
	fmt.Println("Predecessor:")
	n.pNode.mu.Lock()
	n.pNode.n.printNode()
	n.pNode.mu.Unlock()
	fmt.Println()
	fmt.Println("Successors:")
	n.data.mu.Lock()
	for _, s := range n.data.sNodes {
		if s == nil {
			continue
		}
		s.printNode()
		fmt.Println()
	}
	n.data.mu.Unlock()

}

func (n *Node) printNode() {
	fmt.Println("Id: " + n.id.String())
	fmt.Println("Address: " + n.addr)
}

func (n *Node) StoreFile() {

}
func (n *Node) Lookup() {

}
