package rpclib

import (
	"net"
	"net/rpc"
	//"runtime"
)

type Node struct {
	Alived  bool
	Address string
	Client  *rpc.Client
}

func (node *Node) Call(serviceMethod string, args interface{}, reply interface{}) error {
	err := node.Client.Call(serviceMethod, args, reply)
	return err
}

func CreateNode(address string) *Node {
	ret := &Node{}
	ret.Address = address
	conn, err := net.Dial("tcp", address)
	if err != nil {
		ret.Alived = false
		return ret
	}
	ret.Client = rpc.NewClient(conn)
	ret.Alived = true
	return ret
}