FROM ubuntu:latest

RUN apt-get update

RUN apt-get -y install zookeeper bash gosu netcat

COPY init.sh /

COPY zkEnv.sh /usr/share/zookeeper/bin/zkEnv.sh

CMD ["/init.sh"]
