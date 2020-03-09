#!/bin/bash

go run ./main.go kvserv -c ./config/8801.ini &
go run ./main.go kvserv -c ./config/8802.ini &
go run ./main.go kvserv -c ./config/8803.ini &
go run ./main.go kvserv -c ./config/8804.ini &
go run ./main.go kvserv -c ./config/8805.ini &
