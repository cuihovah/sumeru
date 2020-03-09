package raft

import "os"

type Proc interface {
	Do(string, interface{}, interface{}) error
	WriteSnapshot(string, int64) error
	ReadSnapshot(string, *int64) error
	WriteDataIni(*os.File, int64) error
}
