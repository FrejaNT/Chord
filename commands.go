package main

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

func (n *Node) printState() {
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
	for _, s := range n.sNodes.n {
		if s == nil {
			continue
		}
		s.printNode()
		fmt.Println()
	}
	fmt.Println()
	fmt.Println("Fingers:")
	for _, f := range n.finger {
		if f == nil {
			continue
		}
		f.printNode()
		fmt.Println()
	}
}

func (n *Node) printNode() {
	fmt.Println("Id: " + n.id.String())
	fmt.Println("Address: " + n.addr)
}

func (n *Node) storeFile(path string) {
	abs, err := filepath.Abs(path)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	var split []string
	if runtime.GOOS == "windows" {
		split = strings.Split(path, "\"")
	} else {
		split = strings.Split(path, "/")
	}
	name := split[len(split)-1]

	data, err := os.ReadFile(abs)
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	if err := n.sendFile(name, data); err != nil {
		fmt.Println(err.Error())
		return
	}
	fmt.Println("Store Successful")
}

// Tries to find file "name" and prints Host id, -address and content
func (n *Node) lookup(name string) {
	id := getMod(hashString(name))
	n_, err := n.find(id, n)
	if err != nil {
		fmt.Println("could not find host   " + err.Error())
		return
	}
	args := Get_File_Args{Name: name}
	reply := Get_File_Reply{}
	if err := n.dial("Node.GetFile", &args, &reply, n_.addr); err != nil {
		n.removeSuccNode(n_)
		fmt.Println("could not get file   " + err.Error())
		return
	}
	fmt.Println()
	fmt.Println("Host ID: " + n_.id.String())
	fmt.Println("Host Address: " + n_.addr)
	fmt.Println("Contents of " + name + ":")
	fmt.Println(string(reply.Data))
}

// Creates a text file
func (n *Node) createFile(name string, content string) {
	abs, err := filepath.Abs(name)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	f, err := os.Create(abs)
	if err != nil {
		fmt.Println("could not create file   " + err.Error())
		return
	}
	if _, err := f.Write([]byte(content)); err != nil {
		fmt.Println("could not write file   " + err.Error())
		return
	}
	fmt.Println("File created")
}
