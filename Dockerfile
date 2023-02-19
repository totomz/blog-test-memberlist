FROM scratch
WORKDIR /opt/gossip
COPY ./bin/main .

ENTRYPOINT ["/opt/gossip/main"] 