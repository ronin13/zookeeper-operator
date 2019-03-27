FROM ubuntu:latest

ENV OPERATOR=/usr/local/bin/zookeeper-operator \
    USER_UID=1001 \
    USER_NAME=zookeeper-operator

# install operator binary
COPY build/_output/bin/zookeeper-operator ${OPERATOR}

COPY build/bin /usr/local/bin
RUN  /usr/local/bin/user_setup

ENTRYPOINT ["/usr/local/bin/entrypoint"]

USER ${USER_UID}
