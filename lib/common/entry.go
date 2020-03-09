package common

import (
	"container/heap"
	"encoding/json"
)

type Entry struct {
	Id         int64             `json:"id"`
	Time       string            `json:"time"`
	Term       int               `json:"term"`
	Operate    string            `json:"operate"`
	Preference string            `json:"preference"`
	Cmd        string            `json:"cmd"`
	Value      map[string]string `json:"value"`
}

type Entries []Entry

func (e Entry) ToString() string {
	b, _ := json.Marshal(e)
	return string(b)
}

func (eh Entries) Len() int {
	return len(eh)
}

func (eh Entries) Less(i, j int) bool {
	if eh.Len() > i && eh.Len() > j {
		return eh[i].Id < eh[j].Id
	} else {
		return true
	}
}

func (eh Entries) Swap(i, j int) {
	if eh.Len() > i && eh.Len() > j {
		eh[i], eh[j] = eh[j], eh[i]
	}
}

// ====

func (eh *Entries) Push(x interface{}) {
	*eh = append(*eh, x.(Entry))
}

func (eh *Entries) Pop() interface{} {
	old := *eh
	n := len(old)
	x := old[n-1]
	*eh = old[0 : n-1]
	return x
}

func CreateEntries() *Entries {
	hp := &Entries{}
	heap.Init(hp)
	return hp
}
