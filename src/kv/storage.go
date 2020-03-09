package kv

import (
	"../../lib/common"
	"crypto/md5"
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/unknwon/goconfig"
	"io/ioutil"
	"os"
	"path"
	"strconv"
	"strings"
	"sync"
)

type Storage struct {
	data sync.Map
	changeList *sync.Map
}

func hash(str string) string {
	h := md5.New()
	h.Write([]byte(str))
	ret := hex.EncodeToString(h.Sum(nil))
	return ret[0:2]
}

func (s *Storage) WriteDataIni(f *os.File, content int64) error {
	ctx := make([]string, 0)
	ctx = append(ctx, "[base]")
	commitIndex := fmt.Sprintf("commit_index=%d", content)
	ctx = append(ctx, commitIndex)
	f.Write([]byte(strings.Join(ctx, "\n")))
	return nil
}

func (s *Storage) WriteSnapshot(datafile string, index int64) error {
	isChange := false
	s.changeList.Range(func(k interface{}, _ interface{}) bool {
		filename := path.Join(datafile, k.(string))
		result := make([]string, 0)
		isChange = true
		s.data.Range(func(key interface{}, value interface{}) bool {
			if hash(key.(string)) == k.(string) {
				tmp := fmt.Sprintf("%s -> %s", key.(string), value.(string))
				result = append(result, tmp)
			}
			return true
		})
		ioutil.WriteFile(filename, []byte(strings.Join(result, "\n")), 0644)
		return true
	})
	if isChange {
		f, err := os.OpenFile(path.Join(datafile, "data.ini"), os.O_CREATE|os.O_RDWR|os.O_TRUNC, 0644)
		if err != nil {
			return err
		}
		defer f.Close()
		s.WriteDataIni(f, index)
		s.changeList = new(sync.Map)
	}
	return nil
}

func (s *Storage) ReadSnapshot(datafile string, reply *int64) error {
	cfg, err := goconfig.LoadConfigFile(path.Join(datafile, "data.ini"))
	if err != nil {
		return err
	}
	strIndex, err := cfg.GetValue("base", "commit_index")
	commitIndex, _ := strconv.Atoi(strIndex)
	*reply = int64(commitIndex)
	fs, err := ioutil.ReadDir(datafile)
	if err != nil {
		return err
	}
	for _, f := range fs {
		if f.IsDir() || f.Name() == "data.ini"{
			continue
		}
		filepath := path.Join(datafile, f.Name())
		content, err := ioutil.ReadFile(filepath)
		if err != nil {
			continue
		}
		for _, ctx := range strings.Split(string(content), "\n") {
			kv := strings.Split(ctx, "->")
			if len(kv) == 2 {
				s.data.Store(strings.TrimSpace(kv[0]), strings.TrimSpace(kv[1]))
			}
		}
	}
	return nil
}

func (s *Storage) Do(cmd string, value interface{}, reply interface{}) error {
	switch cmd {
	case "PUT":
		kv := value.(map[string]string)
		pcr := &common.RetValue{}
		err := s.put(kv["Key"], kv["Value"], pcr)
		*(reply.(*common.RetValue)) = *pcr
		return err
	case "GET":
		kv := value.(map[string]string)
		key := kv["Key"]
		err := s.get(key, reply.(*common.RetValue))
		return err
	case "SYNC":
		return nil
	default:
		err := errors.New("Not matched")
		return err
	}
}

func (s *Storage) put(key string, value string, reply *common.RetValue) error {
	k := hash(key)
	s.data.Store(key, value)
	s.changeList.Store(k, 1)
	*reply = common.RetValue{
		Code: 0,
		Msg:  "OK",
		Data: true,
	}
	return nil
}

func (s *Storage) get(key string, reply *common.RetValue) error {
	ret, _ := s.data.Load(key)
	*reply = common.RetValue{
		Code: 0,
		Msg:  "OK",
		Data: ret,
	}
	return nil
}

func New() *Storage {
	rt := &Storage{}
	rt.changeList = new(sync.Map)
	return rt
}
