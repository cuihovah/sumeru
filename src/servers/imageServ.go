package servers

import (
	_ "../../lib/common"
	. "../kv"
	"encoding/json"
	"fmt"
	"github.com/julienschmidt/httprouter"
	"github.com/satori/go.uuid"
	"github.com/unknwon/goconfig"
	"io/ioutil"
	"mime"
	"net/http"
	"path"
	"strings"
)

type PostImageData struct {
	Key string `json:"key"`
	Value string `json:"value"`
}

var homePageContent []byte

func PostImage(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	w.Header().Set("Content-Type", "application/json")
	r.ParseForm()
	uid, err := uuid.NewV4()
	uploadFile, handle, err := r.FormFile("image")
	ext := strings.ToLower(path.Ext(handle.Filename))
	if err != nil {
		data := fmt.Sprintf("{\"msg\":\"%s\"}", err.Error())
		w.Write([]byte(data))
		return
	}
	buf, err := ioutil.ReadAll(uploadFile)
	buf = append(buf, []byte("=mime=")...)
	buf = append(buf, []byte(mime.TypeByExtension(ext))...)
	reply, err := sc.CallPut(fmt.Sprintf("%s", uid), string(buf))
	if err != nil {
		data := fmt.Sprintf("{\"msg\":\"%s\"}", err.Error())
		w.Write([]byte(data))
		return
	}
	reply.Data = fmt.Sprintf("%s", uid)
	data, _ := json.Marshal(*reply)
	w.Write([]byte(data))
	return
}

func GetImage(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	uid := ps.ByName("name")
	reply, err := sc.CallGet(uid)
	if err != nil {
		data := fmt.Sprintf("{\"msg\":\"%s\"}", err.Error())
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(data))
	} else {
		buf := reply.Data.(string)
		dt := strings.Split(buf, "=mime=")
		w.Header().Set("Content-Type", dt[1])
		w.Write([]byte(dt[0]))
	}
}

func GetHomePage(w http.ResponseWriter, _ *http.Request, _ httprouter.Params) {
	w.Header().Set("Content-Type", "text/html")
	w.Write(homePageContent)
}

func RunImageServer(filename string) {
	cfg, err := goconfig.LoadConfigFile(filename)
	if err != nil {
		return
	}
	cluster, _ := cfg.GetValue("kvserv", "cluster")
	httpHost, _ := cfg.GetValue("image", "image_host")
	httpPort, _ := cfg.GetValue("image", "image_port")
	homepage, _ := cfg.GetValue("image", "homepage")
	var peers []string
	json.Unmarshal([]byte(cluster), &peers)
	homePageContent, _ = ioutil.ReadFile(homepage)
	fmt.Println(string(homePageContent))
	router := httprouter.New()
	router.POST("/images", PostImage)
	router.GET("/images/:name", GetImage)
	router.GET("/", GetHomePage)
	sc = NewClient(peers)
	fmt.Printf("ImageServe Leader is %s\n", sc.Leader)
	http.ListenAndServe(fmt.Sprintf("%s:%s", httpHost, httpPort),  router)
}