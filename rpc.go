package main

import (
	"crypto/tls"
	"encoding/gob"
	"fmt"
	"log"
	"net/rpc"
)

// create listener
func (n *Node) Server() {
	l, err := tls.Listen("tcp", n.addr, &n.tls)
	if err != nil {
		log.Fatal(err.Error())
	}
	defer l.Close()

	for {
		conn, err := l.Accept()
		if err != nil {
			fmt.Printf("Connection error: %v", err)
			continue
		}
		go rpc.ServeConn(conn)
	}

}

// Dial
func (n *Node) Dial(rpcname string, args interface{}, reply interface{}, addr string) bool {
	conn, err := tls.Dial("tcp", addr, &n.tls)
	if err != nil {
		fmt.Println(err.Error())
		return false
	}
	defer conn.Close()

	client := rpc.NewClient(conn)

	err = client.Call(rpcname, args, reply)
	if err != nil {
		fmt.Println(err.Error())
		return false
	}
	return true
}

type Find_successor_Args struct {
	Id []byte
}
type Find_successor_Reply struct {
	Id   []byte
	Addr string
	OK   bool
}
type Call_Stabilize_Args struct {
}
type Call_Stabilize_Reply struct {
	Id   []byte
	Addr string
	OK   bool
}
type Notify_Args struct {
	Id   []byte
	Addr string
}
type Notify_Reply struct {
	OK bool
}
type Call_predecessor_Args struct {
}
type Call_predecessor_Reply struct {
	OK bool
}

func GobRegister() {
	gob.Register(Find_successor_Args{})
	gob.Register(Find_successor_Reply{})
	gob.Register(Call_Stabilize_Args{})
	gob.Register(Call_Stabilize_Reply{})
	gob.Register(Notify_Args{})
	gob.Register(Notify_Reply{})
	gob.Register(Call_predecessor_Args{})
	gob.Register(Call_predecessor_Reply{})
}
