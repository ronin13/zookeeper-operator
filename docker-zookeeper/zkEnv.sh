ZOOMAIN=org.apache.zookeeper.server.quorum.QuorumPeerMain
ZOOCFGDIR=/etc/zookeeper/conf
ZOOCFG=/etc/zookeeper/conf/zoo.cfg
ZOO_LOG_DIR=/var/log/zookeeper
ZOO_LOG4J_PROP=INFO,CONSOLE
JMXLOCALONLY=true
JAVA_OPTS=""

# If ZooKeeper is started through systemd, this will only be used for command
# line tools such as `zkCli.sh` and not for the actual server
JAVA=/usr/bin/java

# TODO: This is really ugly
# How to find out which jars are needed?
# Seems that log4j requires the log4j.properties file to be in the classpath
CLASSPATH="/etc/zookeeper/conf:/usr/share/java/jline.jar:/usr/share/java/log4j-1.2.jar:/usr/share/java/xercesImpl.jar:/usr/share/java/xmlParserAPIs.jar:/usr/share/java/netty.jar:/usr/share/java/slf4j-api.jar:/usr/share/java/slf4j-log4j12.jar:/usr/share/java/zookeeper.jar"
