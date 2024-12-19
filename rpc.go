package main

import (
	"crypto/tls"
	"encoding/gob"
	"fmt"
	"log"
	"net/rpc"
	"time"
)

// create listener
func (n *Node) server() {
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

// dial
func (n *Node) dial(rpcname string, args interface{}, reply interface{}, addr string) error {
	conn, err := tls.Dial("tcp", addr, &n.tls)
	if err != nil {
		return err
	}
	defer conn.Close()

	conn.SetDeadline(time.Now().Add(2 * time.Second))

	client := rpc.NewClient(conn)
	err = client.Call(rpcname, args, reply)
	if err != nil {
		return err
	}
	return nil
}

// Dials all successors
func (n *Node) dialSuccessors(rpcname string, args interface{}, reply interface{}) error {
	var n_ *Node
	i := 0
	for i < n.r {
		n_ = n.getSuccessor(i)
		if n_.addr == n.addr {
			i++
			continue
		}
		if err := n.dial(rpcname, args, reply, n_.addr); err == nil {
			i++
			continue
		}
		n.removeSuccNode(n_)
		i--
	}
	return nil
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
	Id     []byte
	Addr   string
	SIds   [][]byte
	SAddrs []string
	OK     bool
}
type Call_predecessor_Args struct {
}
type Call_predecessor_Reply struct {
	OK bool
}
type Send_File_Args struct {
	Name string
	Data []byte
}
type Send_File_Reply struct {
}

type Send_Replica_Args struct {
	Names []string
	Data  [][]byte
	R     int
	Rep   []byte
}
type Send_Replica_Reply struct {
}
type Get_File_Args struct {
	Name string
}
type Get_File_Reply struct {
	OK   bool
	Data []byte
}
type Clear_Files_Args struct {
	Rep []byte
}
type Clear_Files_Reply struct {
}
type Get_Files_Args struct {
	Start []byte
	End   []byte
}
type Get_Files_Reply struct {
	Data  [][]byte
	Names []string
}

func gobRegister() {
	gob.Register(Find_successor_Args{})
	gob.Register(Find_successor_Reply{})
	gob.Register(Call_Stabilize_Args{})
	gob.Register(Call_Stabilize_Reply{})
	gob.Register(Notify_Args{})
	gob.Register(Notify_Reply{})
	gob.Register(Call_predecessor_Args{})
	gob.Register(Call_predecessor_Reply{})
	gob.Register(Send_File_Args{})
	gob.Register(Send_File_Reply{})
	gob.Register(Send_Replica_Args{})
	gob.Register(Send_Replica_Reply{})
	gob.Register(Get_File_Args{})
	gob.Register(Get_File_Reply{})
	gob.Register(Clear_Files_Args{})
	gob.Register(Clear_Files_Reply{})
	gob.Register(Get_Files_Args{})
	gob.Register(Get_Files_Reply{})
}
