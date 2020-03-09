package kv

import (
	. "../../lib/common"
	"../../lib/raft"
	"../../lib/rpclib"
	"encoding/gob"
	"errors"
	"sync"
	"time"
)

type StorageClient struct {
	index int
	mu sync.Mutex
	Leader *rpclib.Node
	peers map[string]*rpclib.Node
	cluster chan *rpclib.Node
}

func assign(peers []string) map[string]*rpclib.Node {
	ret := make(map[string]*rpclib.Node)
	for _, i := range peers {
		ret[i] = rpclib.CreateNode(i)
	}
	return ret
}

func (sc *StorageClient) CallPut(key string, value string) (*RetValue, error) {
	reply := &RetValue{}
	if reply == nil {
		return reply, errors.New("reply is nil")
	}
	kv := make(map[string]string)
	kv["Key"] = key
	kv["Value"] = value
	entry := Entry {
		Time: time.Now().String(),
		Operate: "W",
		Cmd: "PUT",
		Value: kv,
	}
	err := sc.Leader.Call("Raft.Exec", entry, reply)
	return reply, err
}

func (sc *StorageClient) CallGet(key string) (*RetValue, error) {
	reply := &RetValue{}
	kv := make(map[string]string)
	kv["Key"] = key
	entry := Entry {
		Time: time.Now().String(),
		Operate: "R",
		Cmd: "GET",
		Value: kv,
	}
	ct := getConnection(sc.cluster)
	err := ct.Call("Raft.Exec", entry, reply)
	return reply, err
}

func init() {
	gob.Register(TopologyData{})
}

func getConnection(ch chan *rpclib.Node) *rpclib.Node {
	var retval *rpclib.Node
	for {
		retval = <- ch
		ch <- retval
		if retval.Alived == true {
			return retval
		}
		time.Sleep(time.Millisecond * Tick)
	}
	return retval
}

func NewClient(peers []string) *StorageClient {
	sc := &StorageClient{}
	sc.cluster = make(chan *rpclib.Node, len(peers))
	sc.peers = assign(peers)
	for _, v := range sc.peers {
		sc.cluster <- v
	}
	ch := make(chan bool)
	closed := false
	go func(){
		for {
			var wg sync.WaitGroup
			for _, v := range sc.peers {
				wg.Add(1)
				go func(node *rpclib.Node){
					if node.Alived == false {
						*node = *rpclib.CreateNode(node.Address)
					} else {
						reply := &HeartBeatRetVal{}
						err := node.Call("Raft.HeartBeat", raft.NoteVotedArgs{"client", "", 0}, &reply)
						if err != nil {
							node.Alived = false
						}
						if reply.State == "leader" {
							sc.Leader = node
							if closed != true {
								ch <- true
							}
						}
					}
					wg.Done()
				}(v)
			}
			wg.Wait()
			time.Sleep(time.Millisecond * Tick)
		}
	}()
	<-ch
	close(ch)
	closed = true
	return sc
}