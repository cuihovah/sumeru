#!/bin/bash

LS=(8801 8802 8803 8804 8805 18801 18802 18803 18804 18805)
# LS=(8801 8802 8803 8804 8805)

# echo $LS
if [ -z ${1} ];then
    for i in ${LS[@]};
    do
        ps -ef | grep "main" | grep ${i} | awk '{print $2}' | xargs kill -9
    done
else
    ps -ef | grep "main" | grep ${1} | awk '{print $2}' | xargs kill -9
fi
