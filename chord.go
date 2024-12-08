package main

import (
	"crypto/sha1"
	"crypto/tls"
	"encoding/binary"
	"fmt"
	"log"
	"net/rpc"
	"slices"
	"strconv"
)

const M = 10

type Node struct {
	id   uint64
	addr string
	tls  tls.Config
	data Data
}
type Data struct {
	kvs    []KeyValue
	finger []Finger
	sNodes []*Node
	pNode  *Node
}

type KeyValue struct {
	Key   int
	Value []byte
}

type maybe struct {
	id   uint64
	addr string
}

type Finger struct {
	i uint64
	n *Node
}

// main func

// init Node & rpc?

// create listener
func (n *Node) server() {
	key, pub, err := GenerateRSA()
	if err != nil {
		log.Fatal(err.Error())
	}
	cert, err := tls.LoadX509KeyPair(pub, key)
	if err != nil {
		log.Fatalf("Failed to load server certificate: %v", err)
	}

	n.tls = tls.Config{Certificates: []tls.Certificate{cert}}

	l, err := tls.Listen("tcp", n.addr, &n.tls)
	if err != nil {
		log.Fatal(err.Error())
	}
	defer l.Close()

	for {
		conn, err := l.Accept()
		if err != nil {
			log.Printf("Connection error: %v", err)
			continue
		}
		go rpc.ServeConn(conn)
	}

}

// Dial
func (n *Node) Dial(rpcname string, args interface{}, reply interface{}, addr string) bool {
	conn, err := tls.Dial("tcp", addr, &n.tls)
	if err != nil {
		log.Fatal(err.Error())
	}
	defer conn.Close()

	client := rpc.NewClient(conn)

	err = client.Call("Node.Find_Successor", &args, &reply)
	if err != nil {
		log.Fatal(err.Error())
	}
	return true
}

func (n *Node) AddSuccNode(new *Node) error {
	if len(n.data.sNodes) == cap(n.data.sNodes) {
		return fmt.Errorf("sNodes is full")
	}
	index := 0
	for _, n_ := range n.data.sNodes {
		if new.id > n_.id {
			index++
		}
	}
	n.data.sNodes = slices.Insert(n.data.sNodes, index, new)
	return nil
}
func (n *Node) RemoveSuccNode(old *Node) error {
	for i, n_ := range n.data.sNodes {
		if old == n_ {
			n.data.sNodes = slices.Delete(n.data.sNodes, i, i+1)
			return nil
		}
	}
	return fmt.Errorf("Node is not a successor")
}

func HashString(s string) uint64 {
	return HashData([]byte(s))
}
func HashInt(i int) uint64 {
	return HashData([]byte(strconv.Itoa(i)))
}
func HashData(d []byte) uint64 {
	hash := sha1.Sum(d)
	res := binary.BigEndian.Uint64(hash[:]) % M
	return res
}
