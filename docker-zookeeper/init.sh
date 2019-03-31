#!/bin/bash 
set -uex

ZOO_CONFIG='/etc/zookeeper/conf/zoo.cfg'
ZOO_DATA_DIR='/var/lib/zookeeper'

if [[ -d $ZOO_DATA_DIR ]];then
    chown -R zookeeper:zookeeper $ZOO_DATA_DIR
fi

myhostname=$(hostname)

myord=$(cut -d- -f2 <<< $myhostname)
htemp=$(cut -d- -f1 <<< $myhostname)
sdomain=$(hostname -f | cut -d. -f2)
myord=$((myord+1))


for id in ${ZOO_IDS//,/ }; do
    tid=$(( id-1 ))
    #if [[ $id -eq $myord ]];then
        #echo "server.${id}=0.0.0.0:2888:3888" >> "$ZOO_CONFIG"
    #else
    echo "server.${id}=${htemp}-${tid}.${sdomain}:2888:3888" >> "$ZOO_CONFIG"
    #fi
done

echo $myord > "$ZOO_DATA_DIR/myid"

sleep 2

exec gosu zookeeper /usr/share/zookeeper/bin/zkServer.sh start-foreground
