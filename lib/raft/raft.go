package raft

import (
	"../common"
	"../rpclib"
	"container/heap"
	"encoding/json"
	"fmt"
	"github.com/unknwon/goconfig"
	"io/ioutil"
	"math/rand"
	"net"
	"net/rpc"
	"os"
	"path"
	"sort"
	"strings"
	"sync"
	"time"
)

type Raft struct {
	term         int
	voteNr       int
	HttpPort     string
	mu           sync.Mutex
	EntryMu      sync.Mutex
	count        int
	Address      string
	Leader       string
	state        string
	timeout      int
	l            net.Listener
	Cluster      []*rpclib.Node
	peers        []string
	Entries      *common.Entries
	entryChannel chan *common.Entry
	EntryFile    *os.File
	EntryPath    string
	OpLogFile    *os.File
	OpLogPath    string
	shutdown     chan bool
	wait         chan bool
	proc         Proc
	andex        int64
	amu          sync.Mutex
	bndex        int64
	bmu          sync.Mutex
	indexLocker  sync.Mutex
	dataPath     string
	Server       *rpc.Server
}

type commonEntry []common.Entry

// sort
func (c commonEntry) Len() int {
	return len(c)
}

func (c commonEntry) Swap(i, j int) {
	c[i], c[j] = c[j], c[i]
}

func (c commonEntry) Less(i, j int) bool {
	return c[i].Id < c[j].Id
}

func (r *Raft) RefactorCluster() {
	r.Cluster = make([]*rpclib.Node, 0, 1024)
	for _, i := range r.peers {
		if i == r.Address {
			continue
		}
		node := rpclib.CreateNode(i)
		if node != nil {
			r.Cluster = append(r.Cluster, node)
		}
	}
}

func (r *Raft) Restore() {
	content, _ := ioutil.ReadFile(r.EntryPath)
	logs := make([]common.Entry, 0, 1000000)
	for _, i := range strings.Split(string(content), "\n") {
		entry := common.Entry{}
		err := json.Unmarshal([]byte(i), &entry)
		if err == nil {
			logs = append(logs, entry)
		}
	}
	sort.Sort(commonEntry(logs))
	for _, entry := range logs {
		r.indexLocker.Lock()
		r.andex = entry.Id
		r.bndex = entry.Id
		r.term = entry.Term
		r.indexLocker.Unlock()
		reply := &common.RetValue{}
		r.proc.Do(entry.Cmd, entry.Value, reply)
	}
	f, err := os.OpenFile(path.Join(r.dataPath, "data.ini"), os.O_CREATE|os.O_RDWR|os.O_TRUNC, 0644)
	if err == nil {
		r.proc.WriteDataIni(f, r.andex)
		f.Close()
	}
}

func (r *Raft) Run() {
	r.startRPCServer()
	go func() {
		for {
			time.Sleep(time.Second * 30)
			r.EntryMu.Lock()
			r.OpLogFile.Close()
			r.OpLogFile, _ = os.OpenFile(r.OpLogPath, os.O_RDWR|os.O_TRUNC, 0644)
			r.proc.WriteSnapshot(r.dataPath, r.bndex)
			r.EntryMu.Unlock()
		}
	}()
	go func() {
		for {
			time.Sleep(time.Millisecond * common.Tick)
			if r.state == Leader {
				r.ImAlived()
			}
		}
	}()
	go func() {
		for {
			// 这个地方逻辑有问题
			if r.Entries.Len() > 0 && (*r.Entries)[0].Id <= r.bndex+1 {
				t := heap.Pop(r.Entries).(common.Entry)
				if t.Id == r.bndex+1 {
					r.bmu.Lock()
					r.bndex = t.Id
					r.bmu.Unlock()
					r.entryChannel <- &t
				}
			} else {
				time.Sleep(common.Tick * 10 * time.Millisecond)
			}
		}
	}()
	go func() {
		for et := range r.entryChannel {
			t := &common.RetValue{0, "OK", false}
			r.EntryMu.Lock()
			if r.state == Follower {
				r.proc.Do(et.Cmd, et.Value, t)
			}
			str, _ := json.Marshal(et)
			str = append(str, '\n')
			r.EntryFile.Write(str)
			r.OpLogFile.Write(str)
			r.EntryMu.Unlock()
		}
	}()
	go func() {
		for {
			time.Sleep(time.Millisecond * common.Tick * 5)
			if r.state == Follower {
				r.mu.Lock()
				r.count++
				r.mu.Unlock()
				if r.count >= r.timeout {
					fmt.Printf("%s is timeout, count: %d and timeout: %d\n", r.Address, r.count, r.timeout)
					r.mu.Lock()
					r.term++
					r.state = Candidate
					r.count = 0
					r.RefactorCluster()
					r.mu.Unlock()
					voteNr := r.SendVoted()
					if voteNr > len(r.peers)/2 {
						r.mu.Lock()
						fmt.Printf("%s is leader\n", r.Address)
						r.state = Leader
						r.Leader = r.Address
						r.mu.Unlock()
						r.SendNotice()
					} else {
						fmt.Printf("%s election is failed\n", r.Address)
						r.timeout = r.timeout + 20
						r.state = Follower
					}
				}
			}
		}
	}()
	r.Wait()
}

func New(filename string, p Proc) *Raft {
	cfg, err := goconfig.LoadConfigFile(filename)
	if err != nil {
		return nil
	}
	host, _ := cfg.GetValue("kvserv", "bind_ip")
	port, _ := cfg.GetValue("kvserv", "port")
	cluster, _ := cfg.GetValue("kvserv", "cluster")
	httpPort, _ := cfg.GetValue("kvserv", "http_port")
	entryPath, _ := cfg.GetValue("kvserv", "entry_path")
	dataPath, _ := cfg.GetValue("kvserv", "data_path")
	opLogPath, _ := cfg.GetValue("kvserv", "oplog_path")
	var peers []string
	json.Unmarshal([]byte(cluster), &peers)
	address := fmt.Sprintf("%s:%s", host, port)
	retval := &Raft{}
	retval.HttpPort = httpPort
	retval.term = 0
	retval.count = 0
	retval.voteNr = 0
	retval.Leader = ""
	retval.Address = address
	retval.peers = peers
	retval.dataPath = dataPath
	retval.proc = p
	retval.state = Follower
	retval.Entries = common.CreateEntries()
	retval.shutdown = make(chan bool)
	retval.wait = make(chan bool)
	rand.Seed(time.Now().UnixNano())
	retval.timeout = rand.Intn(40) + 5
	retval.EntryPath = entryPath
	retval.OpLogPath = opLogPath
	f, _ := os.OpenFile(retval.OpLogPath, os.O_RDWR|os.O_CREATE, 0644)
	retval.OpLogFile = f
	retval.proc.ReadSnapshot(retval.dataPath, &retval.bndex)
	retval.Restore()
	f, _ = os.OpenFile(entryPath, os.O_APPEND|os.O_CREATE|os.O_RDWR, 0644)
	retval.EntryFile = f
	retval.entryChannel = make(chan *common.Entry, 1000000)
	fmt.Printf("%s timeout is %d\n", retval.Address, retval.timeout*common.Tick)
	return retval
}
