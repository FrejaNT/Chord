package main

import (
	"fmt"
	"math/big"
	"os"
	"path/filepath"
)

// This file contains functions related to storing/retrieving/deleting files and related RPC methods

func (n *Node) GetFile(args *Get_File_Args, reply *Get_File_Reply) error {
	id := getMod(hashString(args.Name))
	n.files.mu.Lock()
	fname := n.files.f[id.String()].name
	n.files.mu.Unlock()

	if fname == "" {
		return fmt.Errorf("could not find file")
	}
	data, err := os.ReadFile(fname)
	if err != nil {
		return fmt.Errorf("could not read file")
	}
	reply.Data = data
	reply.OK = true
	return nil
}

// Clears all files that either have an identifier between start and end
// or, if rep is non-nil, any files that are a replica from rep.
func (n *Node) removeFiles(start *big.Int, end *big.Int, rep *big.Int) {
	n.files.mu.Lock()
	for k, v := range n.files.f {
		if between(start, convToBig([]byte(k)), end, true) && v.rep == nil || v.rep != nil && rep != nil && rep.Cmp(v.rep) == 0 {
			p, _ := filepath.Abs(v.name)
			os.Remove(p)
			delete(n.files.f, k)
		}
	}
	n.files.mu.Unlock()
}

func (n *Node) ClearReplicas(args *Clear_Files_Args, reply *Clear_Files_Reply) error {
	n.removeFiles(n.id, n.id, convToBig(args.Rep))
	return nil
}

func (n *Node) GetFiles(args *Get_Files_Args, reply *Get_Files_Reply) error {
	data := make([][]byte, 128)
	names := make([]string, 128)

	index := 0
	n.files.mu.Lock()
	for _, v := range n.files.f {
		if between(convToBig(args.Start), getMod(hashString(v.name)), convToBig(args.End), true) {
			p, _ := filepath.Abs(v.name)
			d, err := os.ReadFile(p)
			if err != nil {
				continue
			}
			data[index] = d
			names[index] = v.name
			index++
		}
	}
	n.files.mu.Unlock()
	reply.Data = data
	reply.Names = names
	return nil
}

func (n *Node) addFile(name string, data []byte, rep *big.Int) error {
	id := getMod(hashString(name))
	f, err := os.Create(name)
	if err != nil {
		return err
	}
	if _, err := f.Write(data); err != nil {
		f.Close()
		return err
	}
	n.files.mu.Lock()
	if rep != nil { // if file is a replica
		n.files.f[id.String()] = Value{name: name, rep: rep}
	} else {
		n.files.f[id.String()] = Value{name: name}
	}
	n.files.mu.Unlock()
	return nil
}

// Adds multiple files to host
func (n *Node) addFiles(names []string, data [][]byte, rep *big.Int) error {
	for i := 0; i < len(names); i++ {
		if names[i] == "" || data[i] == nil {
			continue
		}
		if err := n.addFile(names[i], data[i], rep); err != nil {
			return err
		}
	}
	return nil
}

// send file from client to peer host
func (n *Node) sendFile(name string, data []byte) error {
	id := getMod(hashString(name))
	n_, err := n.find(id, n)
	if err != nil {
		return err
	}
	args := Send_File_Args{Name: name, Data: data}
	reply := Send_File_Reply{}
	if err := n.dial("Node.ReceiveFile", &args, &reply, n_.addr); err != nil {
		n.removeSuccNode(n_)
		return n.sendFile(name, data)
	}
	return nil
}

// RPC call for peer host receiving files
func (n *Node) ReceiveFile(args *Send_File_Args, reply *Send_File_Reply) error {
	if err := n.addFile(args.Name, args.Data, nil); err != nil {
		return fmt.Errorf("%s", "couldn't add file "+args.Name+" to client "+n.addr)
	}

	go n.replicateFile(getMod(hashString(args.Name)))

	return nil
}

// Replicates a single file and sends it to all successors,
// useful when a new file has been received and needs to be replicated
func (n *Node) replicateFile(id *big.Int) error {
	n.files.mu.Lock()
	v := n.files.f[id.String()]
	n.files.mu.Unlock()

	if v.name == "" {
		return fmt.Errorf("file could not be found")
	}

	Names := make([]string, 1)
	Data := make([][]byte, 1)

	Names[0] = v.name
	d, err := os.ReadFile(v.name)
	if err != nil {
		return err
	}
	Data[0] = d

	args := Send_Replica_Args{Names: Names, Data: Data, R: -1, Rep: n.id.Bytes()}
	reply := Send_File_Reply{}

	if err := n.dialSuccessors("Node.ReceiveReplica", &args, &reply); err != nil {
		return fmt.Errorf("%s", v.name+" could not be replicated")
	}
	return nil
}

// Replicates all files and sends them either to a single successor if n_ is non-nil,
// or send them to all successors if n_ is nil. Useful when a node gets a new successor.
func (n *Node) replicateFiles(n_ *Node) error {
	n.files.mu.Lock()
	length := len(n.files.f)
	n.files.mu.Unlock()

	if length == 0 {
		return nil
	}

	Names := make([]string, length)
	Data := make([][]byte, length)
	index := 0

	for _, v := range n.files.f {
		if index >= length { // done
			break
		}
		if v.name == "" || v.rep != nil { // if entry is empty or is already a replica
			continue
		}
		Names[index] = v.name
		d, err := os.ReadFile(v.name)
		if err != nil {
			return err
		}
		Data[index] = d
		index++
	}

	args := Send_Replica_Args{Names: Names, Data: Data, R: -1, Rep: n.id.Bytes()}
	reply := Send_File_Reply{}

	if n_ == nil {
		if err := n.dialSuccessors("Node.ReceiveReplica", &args, &reply); err != nil {
			return fmt.Errorf("files could not be replicated")
		}
		return nil
	}
	if err := n.dial("Node.ReceiveReplica", &args, &reply, n_.addr); err != nil {
		return fmt.Errorf("%s", "files could not be replicated to "+n_.addr)
	}
	return nil
}

// RPC call for receiving and storing a replica from another node
func (n *Node) ReceiveReplica(args *Send_Replica_Args, reply *Send_Replica_Reply) error {
	if err := n.addFiles(args.Names, args.Data, convToBig(args.Rep)); err != nil {
		return err
	}
	return nil
}
