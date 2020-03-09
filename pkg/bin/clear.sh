#!/bin/bash

 LS=(8801 8802 8803 8804 8805)

# echo $LS
if [ -z ${1} ];then
    for i in ${LS[@]};
    do
        cat /dev/null > ./log/${i}.log
        ps -ef | grep "main" | grep ${i} | awk '{print $2}' | xargs kill -9
    done
else
  cat /dev/null > ../log/${1}.log
fi
