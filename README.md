# blog-test-memberlist

A simple implementation example of [hashicorp/memberlist](https://github.com/hashicorp/memberlist), a Go library that manages cluster membership and member failure detection using a gossip based protocol.

Read the full blog post: [https://totomz.github.io/posts/2023-02-18-memberlist/](https://totomz.github.io/posts/2023-02-18-memberlist/)

# Basics
The code in [main.go](./main.go) should be self-explanatory. 

The messages you want to broadcast must implement the [Broadcast](https://pkg.go.dev/github.com/hashicorp/memberlist#Broadcast) interface.
To get messages, you need to provide a [Delegate](https://pkg.go.dev/github.com/hashicorp/memberlist#Delegate) to the **memberlist** configuration

# How to test
Install [shMake](github.com/totomz/shmake),then
```shell
# on the first shell, create the docker network and spawn the first container 
# that will listen on port 3031
docker network create gossip
shmake run --port=3031 --node=localhost

# on a separate shall, start 2 more node
# The ip for the first node can be found by inspecting docker:
#  docker inspect -f '{{range .NetworkSettings.Networks}}{{.IPAddress}}{{end}}' <container id>
shmake run --port=3032 --node=172.20.0.2
shmake run --port=3033 --node=172.20.0.2


```