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
	if n.pNode.n != nil {
		n.pNode.n.printNode()
	}
	fmt.Println()
	fmt.Println("Successors:")
	for _, s := range n.data.sNodes {
		if s == nil {
			continue
		}
		s.printNode()
		fmt.Println()
	}
	fmt.Println()
	fmt.Println("Fingers:")
	for _, s := range n.ft.finger {
		if s.n == nil {
			continue
		}
		s.n.printNode()
		fmt.Println()
	}

}

func (n *Node) printNode() {
	fmt.Println("Id: " + n.id.String())
	fmt.Println("Address: " + n.addr)
}

func (n *Node) StoreFile() {

}
func (n *Node) Lookup() {

}
