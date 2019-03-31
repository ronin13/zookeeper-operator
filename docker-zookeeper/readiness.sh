#!/bin/bash

state=$(echo mntr | nc zoos-serv 2181 | grep zk_server_state)

# Bootstrapping-like
if [[ $? -ne 0 ]];then
    echo mntr | nc localhost 2181
else
    echo mntr | nc localhost 2181 | grep  zk_server_state
fi
