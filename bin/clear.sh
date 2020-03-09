#!/bin/bash

 LS=(8801 8802 8803 8804 8805)

# echo $LS
if [ -z ${1} ];then
    for i in ${LS[@]};
    do
        cat /dev/null > ./log/${i}.log
        cat /dev/null > ./oplog/${i}.log
        rm -rf data/${i}
        mkdir -pv data/${i}
    done
else
  cat /dev/null > ../log/${1}.log
  cat /dev/null > ./oplog/${1}.log
  rm -rf data/${1}
  mkdir -pv data/${1}
fi
