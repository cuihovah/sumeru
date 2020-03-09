#!/bin/bash

for((i=0;i<=1000;i++));
do
  curl -X PUT -d "{\"key\":\"cui${i}\",\"value\":\"cui00${i}\"}" "http://127.0.0.1:8870/set" &
done