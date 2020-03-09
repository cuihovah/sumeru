package main

import (
	"./lib/raft"
	"./src/kv"
	"./src/servers"
	"flag"
	"os"
)

var filename string
var wait chan bool
var flagSet *flag.FlagSet

func main() {
	flagSet := flag.NewFlagSet("start", flag.ExitOnError)
	flagSet.StringVar(&filename, "c", "", "config file path")
	flagSet.Parse(os.Args[2:])
	switch os.Args[1] {
	case "kvserv":
		p := kv.New()
		r := raft.New(filename, p)
		r.Run()
	case "http":
		servers.RunKVServer(filename)
		<-wait
	case "image":
		servers.RunImageServer(filename)
		<-wait
	}
}
