package main

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"
)

func hash(i int) string {
	h := md5.New()
	str := strconv.Itoa(i)
	h.Write([]byte(str))
	ret := hex.EncodeToString(h.Sum(nil))
	return ret[0:6]
}

func httpSet(host string, sBegin string, sEnd string) {
	begin, _ := strconv.Atoi(sBegin)
	end, _ := strconv.Atoi(sEnd)
	for i := begin; i <= end; i++ {
		time.Sleep(time.Millisecond * 10)
		url := fmt.Sprintf("http://%s/set?key=%d&value=%s", host, i, hash(i))
		http.Get(url)
	}
}

func main() {
	host := os.Args[1]
	sBegin := os.Args[2]
	sEnd := os.Args[3]
	httpSet(host, sBegin, sEnd)
}