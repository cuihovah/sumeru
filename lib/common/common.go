package common

type TopologyData struct {
	Leader  string   `json:"leader"`
	Cluster []string `json:"cluster"`
}

type RetValue struct {
	Code int
	Msg  string
	Data interface{}
}

type HeartBeatRetVal struct {
	State string
	Begin int64
	End   int64
}

type KV map[string]string

const (
	Tick = 100
)
