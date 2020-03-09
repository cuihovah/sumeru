# sumeru
Learning MIT-6.824 want to implement k-v server with raft

## Usage

### Start Server

```bash
$ ./run start
```

```bash
$ go run main.go kvserv -c config/8801.ini
```

### Start HTTP Server
```bash
$ go run main.go http -c config/http.ini
```

### Start Image Server
```bash
$ go run main.go image -c config/image.ini
```

### Stop Server
```bash
$ ./run stop 8801
```

### Restart Server
```bash
$ ./run restart
```

### Release
```bash
$ ./run release v0.0.1 darwin
```
output file ***sumeru-darwin-v0.0.1-beta.tar.gz***

### Configure Format
```config
# config.ini

[kvserv]
bind_ip=127.0.0.1
port=8801
cluster=["127.0.0.1:8801", "127.0.0.1:8802", "127.0.0.1:8803", "127.0.0.1:8804", "127.0.0.1:8805"]
entry_path=./log/8801.log
oplog_path=./oplog/8801.log
index_file=./index/8801
data_path=./data/8801
[http]
http_host=127.0.0.1
http_port=8870
[image]
image_host=127.0.0.1
image_port=8700
homepage=./static/html/image.html
```

## Architecture diagram
![Architecture](http://106.12.222.112:18870/images/8017eb65-7235-4114-b058-d371f667f275)

## HTTP Interface

- put value
```bash
$ curl -v -X PUT -d '{"key":"cai", "value":"weiwei"}' "http://127.0.0.1:8870/set"
```

- get value
```bash
$ curl -v "http://127.0.0.1:8870/get?key=cai"
```

## Dependencies
- github.com/unknow/goconfig
- github.com/julienschmidt/httprouter
- github.com/satori/go.uuid

## Features
- Raft
- Data Restore
- oplog
- log
- HTTP Interface
- Synchronization log
- Image Service

## Todos
- Client reconnect
- Node Synchronization Snapshot
- Optimize the election process
- Shard

# LICENSE
MIT LICENSE

## Contributor
- cuihovah（cuihovah@gmail.com）
