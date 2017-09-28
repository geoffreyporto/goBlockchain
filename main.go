package main

import (
	"encoding/json"
	"flag"
	"log"
	"net/http"
	"net/url"
	"reflect"
	"strings"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
)

// Message struct
type Message struct {
	Type uint    `json:"type"`
	Data []Block `json:"data"`
}

var addr = flag.String("addr", "localhost:9001", "http service address")
var initialPeers = flag.String("peers", "", "initial peers with the format host1:port1,host2:port2,...")

// Different message types for Message struct
const (
	MsgTypeQueryLatest    = 0
	MsgTypeQueryAll       = 1
	MsgTypeRespBlockchain = 2
)

// InitHTTPServer initializes the http server and REST api
func InitHTTPServer(router *mux.Router, hub *Hub) {
	flag.Parse()

	// setup REST interface
	router = mux.NewRouter().StrictSlash(true)

	router.HandleFunc("/blocks", GetBlocks).Methods("GET")
	router.HandleFunc("/mineBlock", MineBlock).Methods("POST")
	router.HandleFunc("/peers", GetPeers).Methods("GET")
	router.HandleFunc("/addPeer", AddPeer).Methods("POST")

	// Starts websocket connections for peers
	router.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		serveWs(hub, w, r)
	})

	go ConnectToPeers(*initialPeers)

	log.Printf("Listening on %s", *addr)
	log.Fatal(http.ListenAndServe(*addr, router))
}

// GetBlocks handles the /blocks REST request
func GetBlocks(w http.ResponseWriter, r *http.Request) {
	// Send a copy of this node's blockchain
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(b.blockchain)
}

type BlockData struct {
	Data string `json:"data"`
}

// MineBlock handles the /mineBlock REST post
func MineBlock(w http.ResponseWriter, r *http.Request) {
	// Checks for the block in data field
	var data BlockData
	err := json.NewDecoder(r.Body).Decode(&data)
	if err != nil {
		log.Println("MineBlock: Received block failed to prase(JSON)")
	}

	success := b.GenerateNextBlock(data.Data)
	if success {
		hub.broadcastMsg(RespLatestMsg())
	}
}

// GetPeers handles the /peers REST request
func GetPeers(w http.ResponseWriter, r *http.Request) {
	// Sends list of peers this node is connected to
	peersStr := ""
	for peer, v := range hub.peers {
		if v == true {
			peersStr += peer.conn.RemoteAddr().String()
			peersStr += ","
		}
	}
	//log.Printf("GetPeers: data=%s\n", peersStr)
	var p peerStr
	p.Peer = peersStr
	json.NewEncoder(w).Encode(p)
}

type peerStr struct {
	Peer string `json:"peer"`
}

// AddPeer handles the /addPeer REST post
func AddPeer(w http.ResponseWriter, r *http.Request) {
	// Connect to the peer
	var newPeers peerStr

	err := json.NewDecoder(r.Body).Decode(&newPeers)
	if err != nil {
		log.Println("AddPeer: could not decode peer")
	}
	log.Println(newPeers)
	log.Printf("AddPeer: adding=%s", newPeers.Peer)

	ConnectToPeers(newPeers.Peer)
}

// HandleBlockchainResp
func (m *Message) HandleBlockchainResp() {
	// TODO: check if arrays that are unmarshalled are in the order pre marshalling
	latestBlockReceived := m.Data[len(m.Data)-1]
	latestBlockHeld := *b.GetLatestBlock()

	if latestBlockReceived.Index > latestBlockHeld.Index {
		log.Println("HandleBlockchainResp: blockchain is not the latest")
		if reflect.DeepEqual(latestBlockHeld.Hash, latestBlockReceived.PrevHash) {
			log.Println("newHash and Prevhash matches")
			if b.AddBlock(latestBlockReceived) {
				// append new block to our chain
				log.Println("HandleBlockchainResp: Added block")
				hub.broadcastMsg(RespLatestMsg())
			} else {
				log.Println("HandleBlockchainResp: AddBlock failed")
			}
		} else if len(m.Data) == 1 {
			// query chain from peer
			log.Println("HandleBlockchainResp: len 1")
			hub.broadcastMsg(QueryAllMsg())
		} else {
			// received blockchain is longer, therefore replace held chain
			log.Println("HandleBlockchainResp: replacing chain")
			b.ReplaceChain(hub, m.Data)
		}
	} else {
		log.Println("HandleBlockchainResp: do nothing.received chain shorter")
	}
}

// QueryChainLengthMsg generates a Message with type QueryLatest
func QueryChainLengthMsg() *Message {
	return &Message{MsgTypeQueryLatest, []Block{}}
}

// QueryAllMsg generates a Message with type QueryAll
func QueryAllMsg() *Message {
	return &Message{MsgTypeQueryAll, []Block{}}
}

// RespChainMsg generates a Message with type RespBlockchain and data of the full blockchain
func RespChainMsg() *Message {
	return &Message{MsgTypeRespBlockchain, b.blockchain}
}

// RespLatestMsg generates a message with the latest block
func RespLatestMsg() *Message {
	block := make([]Block, 1)
	block[0] = *b.GetLatestBlock()
	return &Message{MsgTypeRespBlockchain, block}
}

// ConnectToPeers connects to the peers' addr:port by first parsing the string with the format
// addr1:port1, addr2:port2,...
func ConnectToPeers(newPeers string) {
	peers := strings.Split(newPeers, ",")
	log.Println(peers)
	log.Println(len(peers))
	if peers[0] == "" {
		return
	}
	for _, peer := range peers {
		// connect to peer
		u := url.URL{Scheme: "ws", Host: peer, Path: "/ws"}
		log.Printf("Connecting to %s", u.String())

		c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
		if err != nil {
			log.Fatal("dial:", err)
		}
		addWs(hub, c)
	}
}

var b *Blockchain
var hub *Hub

func main() {
	// Create the blockchain for this node
	b = new(Blockchain)
	b.Init()

	// Setup hub for websocket connectsion
	hub = newHub()
	go hub.run()

	var router *mux.Router
	InitHTTPServer(router, hub)
}
