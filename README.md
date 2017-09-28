# goBlockchain
### Motivation
A simple blockchain written in Golang to understand the basics of cryptocurrencies. Inspired by [Ihartikk's implmentation](https://github.com/lhartikk/naivechain)

### Usage

Setting up 2 nodes locally

- node 1
```bash
go build
./goBlockchain
```
- node 2
```bash
./goBlockchain -addr ":9002" -peers "localhost:9001"
```
The second node will attemp to connect to node 1

### REST API

Example commands to interact with the REST API

```bash
# Query for the full blockchain
curl http://localhost:9001/blocks

# Mine a block
curl -H "Content-type:application/json" --data '{"data" : "Block Data."}' http://localhost:9001/mineBlock

# Connect to another node
curl -H "Content-type:application/json" --data '{"peer" : "localhost:9002"}' http://localhost:9001/addPeer

# Query for connected nodes
curl http://localhost:9001/peers
```

### Improvements for the future
- Add proof of work algorithm for consensus
- Use binary data over websocket instead of JSON for efficiency
- Docker file to launch and test more easily
