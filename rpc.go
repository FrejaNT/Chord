package main

type Find_successor_Args struct {
	Id uint64
}
type Find_successor_Reply struct {
	Id   uint64
	Addr string
}
type Call_Stabilize_Args struct {
}
type Call_Stabilize_Reply struct {
	Id   uint64
	Addr string
}
type Notify_Args struct {
	Id   uint64
	Addr string
}
type Notify_Reply struct {
}
type Call_predecessor_Args struct {
}
type Call_predecessor_Reply struct {
	OK bool
}
