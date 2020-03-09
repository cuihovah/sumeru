package servers

import (
	_ "../../lib/common"
	. "../kv"
	"encoding/json"
	"fmt"
	"github.com/julienschmidt/httprouter"
	"github.com/unknwon/goconfig"
	"io/ioutil"
	"net/http"
)

var sc *StorageClient

type PutData struct {
	Key string `json:"key"`
	Value string `json:"value"`
}

func PutKey(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	w.Header().Set("Content-Type", "application/json")
	buf, err := ioutil.ReadAll(r.Body)
	if err != nil {
		data := fmt.Sprintf("{\"msg\":\"%s\"}", err.Error())
		w.Write([]byte(data))
		return
	}
	body := PutData{}
	err = json.Unmarshal(buf, &body)
	if err != nil {
		data := fmt.Sprintf("{\"msg\":\"%s\"}", err.Error())
		w.Write([]byte(data))
		return
	}
	reply, err := sc.CallPut(body.Key, body.Value)
	if err != nil {
		data := fmt.Sprintf("{\"msg\":\"%s\"}", err.Error())
		w.Write([]byte(data))
		return
	}
	data, _ := json.Marshal(*reply)
	w.Write([]byte(data))
	return
}

func GetKey(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	w.Header().Set("Content-Type", "application/json")
	query := r.URL.Query()
	reply, err := sc.CallGet(query["key"][0])
	if err != nil {
		data := fmt.Sprintf("{\"msg\":\"%s\"}", err.Error())
		w.Write([]byte(data))
	} else {
		data, _ := json.Marshal(*reply)
		w.Write([]byte(data))
	}
}

func RunKVServer(filename string) {
	cfg, err := goconfig.LoadConfigFile(filename)
	if err != nil {
		return
	}
	cluster, _ := cfg.GetValue("kvserv", "cluster")
	httpHost, _ := cfg.GetValue("http", "http_host")
	httpPort, _ := cfg.GetValue("http", "http_port")
	var peers []string
	json.Unmarshal([]byte(cluster), &peers)
	router := httprouter.New()
	router.PUT("/set", PutKey)
	router.GET("/get", GetKey)
	sc = NewClient(peers)
	fmt.Printf("KVSERV Leader is %s\n", sc.Leader.Address)
	http.ListenAndServe(fmt.Sprintf("%s:%s", httpHost, httpPort),  router)
}