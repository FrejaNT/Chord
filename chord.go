package main

import (
	"bufio"
	"crypto/tls"
	"fmt"
	"log"
	"math/big"
	"net/rpc"
	"os"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"
)

const M = 32

type Node struct {
	id    *big.Int
	addr  string
	tls   tls.Config
	pNode PNode
	data  Data
	ft    FTable
}
type Data struct {
	mu     sync.Mutex
	kvs    []KeyValue
	sNodes []*Node
	r      int
}

type KeyValue struct {
	Key   *big.Int
	Value []byte
}

type PNode struct {
	mu sync.Mutex
	n  *Node
}
type FTable struct {
	mu     sync.Mutex
	finger []Finger
}

type Finger struct {
	i *big.Int
	n *Node
}

// main func
func main() {
	fmt.Println("argie")
	args := os.Args
	n := Node{}
	join, jaddr, ts, tff, tcp := n.initNode(args)
	fmt.Println("initie")
	n.LoadTLS()
	GobRegister()
	rpc.Register(&n)
	go n.Server()
	fmt.Println("servie")
	if join {
		jn := Node{addr: jaddr}
		go n.Join(&jn)
		fmt.Println("joinie")
	}
	if !join {
		fmt.Println("creatie")
		n.Create()
	}

	go n.RepeatNodeFunction(ts, n.Stabilize)
	fmt.Println(tff)
	//go n.RepeatNodeFunction(tff, n.Fix_fingers)
	go n.RepeatNodeFunction(tcp, n.Check_predecessor)

	running := true
	fmt.Println("runnie")
	for running {
		reader := bufio.NewReader(os.Stdin)
		fmt.Print("Enter text: ")
		input, _ := reader.ReadString('\n')

		if runtime.GOOS == "windows" {
			input = strings.TrimRight(input, "\r\n")
		} else {
			input = strings.TrimRight(input, "\n")
		}

		if input == "Lookup" {

		}
		if input == "StoreFile" {

		}
		if input == "PrintState" {
			n.PrintState()
		}
		if input == "q" {
			running = false
		}
	}
	os.Exit(0)
}

// init Node & rpc?
func (n *Node) initNode(args []string) (bool, string, string, string, string) {
	var a, p, ja, jp, ts, tff, tcp, r bool
	var ip, port, jip, jport, its, itff, itcp string
	for i, s := range args {
		fmt.Println(s)
		if i >= len(args)-1 {
			break
		}
		if s == "-a" {
			a = true
			ip = args[i+1]
		}
		if s == "-p" {
			p = true
			port = args[i+1]
		}
		if s == "--ja" {
			ja = true
			jip = args[i+1]
		}
		if s == "--jp" {
			jp = true
			jport = args[i+1]
		}
		if s == "--ts" {
			ts = true
			its = args[i+1] + "ms"

		}
		if s == "--tff" {
			tff = true
			itff = args[i+1] + "ms"
		}
		if s == "--tcp" {
			tcp = true
			itcp = args[i+1] + "ms"
		}
		if s == "-r" {
			j, err := strconv.Atoi(args[i+1])
			if err == nil {
				r = true
				n.data.r = j
				n.data.sNodes = make([]*Node, j)
			}
		}
		if s == "-i" {

		}
	}
	if !(a && p && ts && tff && tcp && r) || ((ja && !jp) || (!ja && jp)) {
		log.Fatal("argie problem")
	}
	n.addr = ip + ":" + port
	n.id = HashString(n.addr)
	if ja && jp {
		return true, jip + ":" + jport, its, itff, itcp
	}
	return false, "", its, itff, itcp

}

func (n *Node) AddSuccNode(new *Node) error {
	n.data.mu.Lock()
	index := 0

	for _, s := range n.data.sNodes {
		if s == nil || Between(n.id, new.id, s.id, false) {
			continue
		}
		if s == new {
			n.data.mu.Unlock()
			return nil
		}
		index++
	}
	if index >= n.data.r {
		n.data.mu.Unlock()
		return nil
	}
	n.data.sNodes[index] = new
	n.data.mu.Unlock()
	return nil
}
func (n *Node) RemoveSuccNode(old *Node) {
	n.data.mu.Lock()
	nils := 0
	for i, n_ := range n.data.sNodes {
		if old == n_ {
			n.data.sNodes[i] = nil
		}
		if n.data.sNodes[i] == nil {
			nils++
		}
	}
	if nils == n.data.r {
		n.AddSuccNode(n)
	}
	n.data.mu.Unlock()
}
func (n *Node) GetSuccNode() *Node {
	n.data.mu.Lock()
	var closest *Node
	for _, n_ := range n.data.sNodes {
		if closest == nil || n_.id.Cmp(closest.id) < 0 {
			closest = n_
		}
	}
	cid := new(big.Int) // trying to be safe
	*cid = *closest.id
	n.data.mu.Unlock()
	return &Node{id: cid, addr: closest.addr}
}

func (n *Node) RepeatNodeFunction(t string, fn func()) {
	t_, err := time.ParseDuration(t)
	if err != nil {
		log.Fatal("tcp not working")
	}
	for {
		fn()
		time.Sleep(t_)
	}
}
