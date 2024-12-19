package main

import (
	"bufio"
	"crypto/tls"
	"flag"
	"fmt"
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
	id     *big.Int
	addr   string
	tls    tls.Config
	pNode  PNode
	sNodes sNodes
	index  int
	finger []*Node
	r      int
	files  Data
	enc    map[string][]byte
}
type sNodes struct {
	mu sync.Mutex
	n  []*Node
}
type Data struct {
	mu sync.Mutex
	f  map[string]Value
}
type Value struct {
	name string
	rep  *big.Int
}
type PNode struct {
	mu sync.Mutex
	n  *Node
}

// main func
func main() {
	n := Node{}
	join, jaddr, ts, tff, tcp := n.initNode()
	n.loadTLS()
	gobRegister()
	rpc.Register(&n)
	go n.server()
	if join {
		jn := Node{addr: jaddr}
		go n.join(&jn)
	}
	if !join {
		n.create()
	}
	go n.repeatNodeFunction(ts, n.stabilize)
	go n.repeatNodeFunction(tff, n.fix_fingers)
	go n.repeatNodeFunction(tcp, n.checkPredecessor)

	running := true

	for running {
		reader := bufio.NewReader(os.Stdin)
		fmt.Print("Enter text: ")
		input, _ := reader.ReadString('\n')

		if runtime.GOOS == "windows" {
			input = strings.TrimRight(input, "\r\n")
		} else {
			input = strings.TrimRight(input, "\n")
		}
		inputs := strings.Split(strings.TrimSpace(input), " ")
		if len(inputs) < 1 && len(inputs) > 3 {
			fmt.Println("Invalid input")
			continue
		}
		fmt.Println(inputs[0])

		if inputs[0] == "Lookup" && len(inputs) == 2 {
			n.lookup(inputs[1])
			continue
		}
		if inputs[0] == "StoreFile" && len(inputs) == 2 {
			n.storeFile(inputs[1])
			continue
		}
		if inputs[0] == "PrintState" {
			n.printState()
			continue
		}
		if inputs[0] == "Create" && len(inputs) == 3 {
			n.createFile(inputs[1], inputs[2])
			continue
		}

		if inputs[0] == "q" {
			running = false
			continue
		}
		fmt.Println("invalid input")
	}
	os.Exit(0)
}

// Initiates the node using the flags and returns Join, join ip, ts, tff, tcp
func (n *Node) initNode() (bool, string, int, int, int) {
	// required flags
	required := []string{"a", "p", "ts", "tff", "tcp", "r"}

	ip := flag.String("a", "error", "Client IP")
	port := flag.Int("p", 0, "Client Port")
	jip := flag.String("ja", "0", "IP to connect to")
	jport := flag.Int("jp", 0, "Port to connect to")
	ts := flag.Int("ts", 0, "ms between invocations of sabilize. [1, 60000]")
	tff := flag.Int("tff", 0, "ms between invocations of fix fingers [1, 60000]")
	tcp := flag.Int("tcp", 0, "ms between invocations of check predecessor")
	r := flag.Int("r", 0, "Number of replica nodes")

	flag.Parse()
	join := false

	seen := make(map[string]bool)
	flag.Visit(func(f *flag.Flag) { seen[f.Name] = true })

	if seen["ja"] || seen["jp"] {
		required = append(append(required, "jp"), "ja")
		join = true
	}
	for _, req := range required {
		if !seen[req] {
			fmt.Fprintf(os.Stderr, "missing required -%s argument/flag\n", req)
			os.Exit(2)
		}
	}
	// Check that values are in range
	err := checkRanges(*ts, *tff, *tcp, *r)
	if err != nil {
		fmt.Println(err)
		os.Exit(2)
	}
	// intialize fields in Node n
	n.addr = *ip + ":" + strconv.Itoa(*port)
	n.id = getMod(hashString(n.addr))
	n.finger = make([]*Node, M)
	n.index = 0
	n.files.f = make(map[string]Value)
	n.enc = make(map[string][]byte)
	n.r = *r
	n.sNodes.n = make([]*Node, *r)

	if join {
		return true, *jip + ":" + strconv.Itoa(*jport), *ts, *tff, *tcp
	}
	return false, "", *ts, *tff, *tcp
}

func checkRanges(ts int, tff int, tcp int, r int) error {
	errmsg := []string{}
	if ts < 1 || ts > 60000 {
		errmsg = append(errmsg, strconv.Itoa(ts)+" not in range [1, 60000]. ")
	}
	if tff < 1 || tff > 60000 {
		errmsg = append(errmsg, strconv.Itoa(tff)+" not in range [1, 60000]. ")

	}
	if tcp < 1 || tcp > 60000 {
		errmsg = append(errmsg, strconv.Itoa(tcp)+" not in range [1, 60000]. ")
	}

	if r < 1 || r > 32 {
		errmsg = append(errmsg, strconv.Itoa(r)+" not in range [1, 32]. ")
	}
	if len(errmsg) == 0 {
		return nil
	}
	return fmt.Errorf("%v", errmsg)
}

// adds successor nodes
func (n *Node) addSuccNodeUnsafe(new *Node) bool {
	if new.id == nil || new.addr == "" {
		return false
	}
	// New node is the same as self or if immidiate successors is nil
	// make self sucecssor
	if n.addr == new.addr && n.sNodes.n[0] == nil {
		n.sNodes.n[0] = n
		return false
	}

	current := new
	added := false
	removed := n.sNodes.n[n.r-1]

	// check if new node should be in successor list
	for i := 0; i < n.r; i++ {
		n_ := n.sNodes.n[i]
		if n_ == nil && n.addr != current.addr {
			n.sNodes.n[i] = current
			added = true
			break
		}
		if n_ == nil || n_.addr == current.addr {
			return false
		}
		// sorts the new node into the list if it is a closer successor
		// than any of the current successors
		if between(n.id, current.id, n_.id, false) {
			temp := n_
			n.sNodes.n[i] = current
			current = temp
			added = true
		}
	}
	// clears file replicas from removed node
	if removed != nil && removed.addr != n.sNodes.n[n.r-1].addr {
		args_ := Clear_Files_Args{n.id.Bytes()}
		reply_ := Clear_Files_Reply{}
		go n.dial("Node.ClearReplicas", &args_, &reply_, removed.addr)
	}
	// replicate files at new successor
	if added {
		go n.replicateFiles(new)
	}
	return added
}

// wrapper to make access to Successor nodes mutually exclusive
func (n *Node) addSuccNode(new *Node) {
	n.sNodes.mu.Lock()
	n.addSuccNodeUnsafe(new)
	n.sNodes.mu.Unlock()
}

// attempts to add the successors supplied
func (n *Node) addSuccessors(sIds [][]byte, sAddrs []string) {
	if len(sAddrs) < 1 {
		return
	}
	n.sNodes.mu.Lock()
	for i := range sIds {
		n.addSuccNodeUnsafe(&Node{id: convToBig(sIds[i]), addr: sAddrs[i]})
	}
	// If n doesn't have r successors the furthest successor is asked for successors
	if len(n.sNodes.n) < n.r-1 {
		n.sNodes.mu.Unlock()
		furthest := n.getFurthestSucc()
		args := Notify_Args{Id: n.id.Bytes(), Addr: n.addr}
		reply := Notify_Reply{}
		if err := n.dial("Node.Notify", &args, &reply, furthest.addr); err != nil {
			n.removeSuccNode(furthest)
			return
		}
		n.addSuccessors(reply.SIds, reply.SAddrs)
		return
	}
	n.sNodes.mu.Unlock()
}

// removes successor from node
func (n *Node) removeSuccNode(old *Node) {
	n.sNodes.mu.Lock()
	nils := 0
	removed := false
	for i := 0; i < n.r; i++ {
		if n.sNodes.n[i] == nil {
			nils++
			continue
		}
		if removed {
			n.sNodes.n[i-1] = n.sNodes.n[i]
			n.sNodes.n[i] = nil
			nils++
			continue
		}
		if old.addr == n.sNodes.n[i].addr {
			n.sNodes.n[i] = nil
			removed = true
			nils++
		}
	}
	// if removed clear replicated files
	if removed {
		args := Clear_Files_Args{n.id.Bytes()}
		reply := Clear_Files_Reply{}
		go n.dial("Node.ClearReplicas", &args, &reply, old.addr)
	}
	if nils >= n.r-1 {
		n.addSuccNodeUnsafe(n)
	}
	n.sNodes.mu.Unlock()
}

// Returns successor at position i in successor list (Node or nil)
func (n *Node) getSuccessor(i int) *Node {
	n.sNodes.mu.Lock()
	succ := n.sNodes.n[i]
	if succ == nil {
		n.sNodes.mu.Unlock()
		return n
	}
	sid := new(big.Int) // trying to be safe
	*sid = *succ.id
	n.sNodes.mu.Unlock()
	return &Node{id: sid, addr: succ.addr}
}

// Returns the successor last in the successor list
func (n *Node) getFurthestSucc() *Node {
	n.sNodes.mu.Lock()

	var furthest *Node
	for i := n.r - 1; i >= 0; i-- {
		n_ := n.sNodes.n[i]
		if n_ != nil {
			furthest = n_
			break
		}
	}

	if furthest == nil {
		n.sNodes.mu.Unlock()
		return n
	}
	fid := new(big.Int)
	*fid = *furthest.id
	n.sNodes.mu.Unlock()
	return &Node{id: fid, addr: furthest.addr}
}

// repeat function which runs the supplied function at an intervall of t ms
func (n *Node) repeatNodeFunction(t int, fn func()) {
	t_ := time.Duration(t) * time.Millisecond
	for {
		fn()
		time.Sleep(t_)
	}
}
