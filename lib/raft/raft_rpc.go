package raft

import (
	"../common"
	"../rpclib"
	"container/heap"
	"encoding/gob"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"net/rpc"
	"strings"
	"sync"
)

type NoteVotedArgs struct {
	Role    string
	Address string
	Term    int
}

type RequestVoteArgs struct {
	Term    int
	Index   int64
	Address string
}

func (r *Raft) startRPCServer() error {
	l, err := net.Listen("tcp", r.Address)
	if err != nil {
		return err
	}
	r.l = l
	r.Server = rpc.NewServer()
	err = r.Server.Register(r)
	if err != nil {
		return err
	}
	go func() {
		r.Server.Accept(r.l)
	}()
	fmt.Printf("rpc server %s is start...\n", r.Address)
	return nil
}

func (r *Raft) stopRPCServer() {
	r.shutdown <- true
}

func (r *Raft) Wait() {
	<-r.wait
}

func (r *Raft) NoteVoted(args RequestVoteArgs, reply *bool) error {
	r.mu.Lock()
	r.Leader = args.Address
	r.term = args.Term
	r.count = 0
	r.Cluster = nil
	fmt.Printf("--- Election %s leader is %s ---\n", r.Address, r.Leader)
	*reply = true
	r.mu.Unlock()
	return nil
}

func (r *Raft) InstallSnapShot(args []common.Entry, reply *bool) error {
	for _, e := range args {
		r.WriteEntries(&e)
	}
	*reply = true
	return nil
}

func (r *Raft) sendEntries(node *rpclib.Node, begin int64, end int64) {
	content, _ := ioutil.ReadFile(r.EntryPath)
	logs := make([]common.Entry, 0, 1000000)
	for _, i := range strings.Split(string(content), "\n") {
		entry := common.Entry{}
		err := json.Unmarshal([]byte(i), &entry)
		if err == nil && entry.Id >= begin && entry.Id <= end {
			logs = append(logs, entry)
		}
	}
	var reply bool
	node.Call("Raft.InstallSnapShot", logs, &reply)
}

func (r *Raft) ImAlived() {
	var wg sync.WaitGroup
	for _, i := range r.Cluster {
		if i.Alived == false {
			p := rpclib.CreateNode(i.Address)
			i.Client = p.Client
			i.Alived = p.Alived
			fmt.Printf("%s reconnect with %s is %t\n", r.Address, i.Address, i.Alived)
		}
		if i.Alived == false {
			continue
		}
		wg.Add(1)
		go func(node *rpclib.Node) {
			reply := &common.HeartBeatRetVal{}
			if node.Alived == false {
				*node = *rpclib.CreateNode(node.Address)
			} else {
				err := node.Call("Raft.HeartBeat", NoteVotedArgs{Leader, r.Address, r.term}, &reply)
				if err != nil {
					node.Alived = false
					fmt.Println(err.Error())
					return
				}
				if reply.Begin < r.andex {
					if r.bndex == -1 {
						r.sendEntries(node, reply.Begin, r.andex)
					} else {
						r.sendEntries(node, reply.Begin, reply.End)
					}
				}
			}
			wg.Done()
		}(i)
	}
	wg.Wait()
}

func (r *Raft) HeartBeat(args NoteVotedArgs, reply *common.HeartBeatRetVal) error {
	if r.Entries.Len() > 0 {
		*reply = common.HeartBeatRetVal{
			State: r.state,
			Begin: r.andex,
			End:   (*r.Entries)[0].Id,
		}
	} else {
		*reply = common.HeartBeatRetVal{
			State: r.state,
			Begin: r.andex,
			End:   -1,
		}
	}
	if args.Role == "client" {
		return nil
	}
	r.mu.Lock()
	if r.Leader == "" {
		r.Leader = args.Address
		r.term = args.Term
		fmt.Printf("Election %s leader is %s with start\n", r.Address, r.Leader)
		r.mu.Unlock()
		return nil
	}
	if r.Leader != args.Address {
		r.mu.Unlock()
		return nil
	}
	if r.state != Follower {
		r.mu.Unlock()
		return nil
	}
	r.count = 0
	r.mu.Unlock()
	return nil
}

func (r *Raft) RequestVoted(args RequestVoteArgs, reply *bool) error {
	if r.state != Follower {
		*reply = false
		return nil
	} else if args.Term < r.term {
		*reply = false
	} else if args.Index < r.andex {
		*reply = false
	} else {
		*reply = true
	}
	r.mu.Lock()
	r.count = 0
	r.mu.Unlock()
	return nil
}

func (r *Raft) SendNotice() error {
	var wg sync.WaitGroup
	r.mu.Lock()
	for _, i := range r.Cluster {
		if i.Alived == false {
			continue
		}
		wg.Add(1)
		go func(node *rpclib.Node) {
			var res bool
			node.Call("Raft.NoteVoted", NoteVotedArgs{Leader, r.Address, r.term}, &res)
			wg.Done()
		}(i)
	}
	wg.Wait()
	r.mu.Unlock()
	return nil
}

func (r *Raft) SendVoted() int {
	var wg sync.WaitGroup
	voteNr := 1
	var mu sync.Mutex
	for _, i := range r.Cluster {
		if i.Alived == false {
			continue
		}
		wg.Add(1)
		go func(node *rpclib.Node) {
			var res bool
			err := node.Call("Raft.RequestVoted", RequestVoteArgs{r.term, r.andex, r.Address}, &res)
			if err == nil && res == true {
				mu.Lock()
				voteNr++
				mu.Unlock()
			} else {
				if err != nil {
					fmt.Printf("%s rejected vote %s\n", node.Address, err.Error())
				} else {
					fmt.Printf("%s rejected vote\n", node.Address)
				}
			}
			wg.Done()
		}(i)
	}
	wg.Wait()
	return voteNr
}

func (r *Raft) WriteEntries(et *common.Entry) error {
	heap.Push(r.Entries, *et)
	return nil
}

func (r *Raft) AppendEntries(et common.Entry, reply *bool) error {
	err := r.WriteEntries(&et)
	if err != nil {
		*reply = false
		return err
	}
	*reply = true
	return nil
}

func (r *Raft) Do(et common.Entry, reply *common.RetValue) error {
	rr := &common.RetValue{}
	err := r.proc.Do(et.Cmd, et.Value, rr)
	if err == nil {
		*reply = *rr
	}
	return err
}

func (r *Raft) Exec(et common.Entry, reply *common.RetValue) error {
	if r.state == Follower && et.Operate == "W" {
		*reply = common.RetValue{
			Code: 2,
			Msg:  "Rejected",
			Data: `The follower node is not allowed to perform "W" operations`,
		}
		return nil
	}
	if et.Operate == "R" {
		rr := &common.RetValue{}
		r.proc.Do(et.Cmd, et.Value, rr)
		*reply = *rr
		return nil
	}
	var wg sync.WaitGroup
	var mu sync.Mutex
	count := 1
	r.indexLocker.Lock()
	et.Id = r.andex + 1
	r.amu.Lock()
	r.andex += 1
	r.amu.Unlock()
	et.Term = r.term
	r.indexLocker.Unlock()
	for _, i := range r.Cluster {
		if i.Alived == false {
			continue
		}
		wg.Add(1)
		go func(node *rpclib.Node) {
			var ret bool
			err := node.Call("Raft.AppendEntries", et, &ret)
			if err == nil && ret == true {
				mu.Lock()
				count++
				mu.Unlock()
			}
			wg.Done()
		}(i)
	}
	wg.Wait()
	if count > len(r.peers)/2 {
		r.WriteEntries(&et)
		rr := &common.RetValue{}
		r.proc.Do(et.Cmd, et.Value, rr)
		*reply = *rr
	}
	return nil
}

func init() {
	gob.Register(common.TopologyData{})
}

func (r *Raft) Topology(_ common.Entry, reply *common.RetValue) error {
	topology := common.TopologyData{
		Leader:  r.Leader,
		Cluster: r.peers,
	}
	ret := common.RetValue{
		Code: 0,
		Msg:  "OK",
		Data: topology,
	}
	*reply = ret
	return nil
}

func (r *Raft) SyncEntries(nextIndex int64, reply *[]common.Entry) error {
	content, _ := ioutil.ReadFile(r.EntryPath)
	rt := make([]common.Entry, 0, 1000000)
	for _, i := range strings.Split(string(content), "\n") {
		entry := common.Entry{}
		err := json.Unmarshal([]byte(i), &entry)
		if err == nil && entry.Id > nextIndex {
			rt = append(rt, entry)
		}
	}
	*reply = rt
	return nil
}
