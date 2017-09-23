package main

import (
	"encoding/json"
	"flag"
	"log"
	"net/http"
	"net/url"
	"strings"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
)

// Message struct
type Message struct {
	Type uint    `json:"type"`
	Data []Block `json:"data"`
}

var addr = flag.String("addr", ":9001", "http service address")

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

	log.Fatal(http.ListenAndServe(*addr, router))
}

// GetBlocks handles the /blocks REST request
func GetBlocks(w http.ResponseWriter, r *http.Request) {
	// Send a copy of this node's blockchain
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(b.blockchain)
}

// MineBlock handles the /mineBlock REST post
func MineBlock(w http.ResponseWriter, r *http.Request) {
	// Checks for the block in data field
	//params := mux.Vars(r)
	var block Block
	err := json.NewDecoder(r.Body).Decode(&block)
	if err != nil {
		log.Println("MineBlock: Received block failed to prase(JSON)")
	}
	success := b.AddBlock(block)
	if success {
		hub.broadcastMsg(RespLatestMsg())
	}
}

// GetPeers handles the /peers REST request
func GetPeers(w http.ResponseWriter, r *http.Request) {
	// Sends list of peers this node is connected to
	log.Println("----Peers----")
	peersStr := ""
	for peer, v := range hub.peers {
		if v == true {
			log.Println(peer.conn.RemoteAddr().String())
			peersStr += peer.conn.RemoteAddr().String()
			peersStr += ","
		}
	}
	json.NewEncoder(w).Encode(peersStr)
}

// AddPeer handles the /addPeer REST post
func AddPeer(w http.ResponseWriter, r *http.Request) {
	// Connect to the peer
	//params := mux.Vars(r)
	var newPeers string

	err := json.NewDecoder(r.Body).Decode(&newPeers)
	if err != nil {
		log.Println("AddPeer: could not decode peer")
	}

	ConnectToPeers(newPeers)
}

/*
// HandleMessages handles the broadcast of messages over websocket
func HandleMessages() {
	for {
		// Grab the next message from the broadcast channel
		msg := <-broadcast
		// Send it out to every client that is currently connected
		for peer := range sockets {
			switch msg.Type {
			case MsgTypeQueryAll:
				// TODO: write response chain message
			case MsgTypeQueryLatest:
				// TODO: write response latest msg
			case MsgTypeRespBlockchain:
				// TODO: handleBlockchainResponse
			default:
				log.Fatal("Unknown message type")
			}
			err := peer.WriteJSON(msg)
			if err != nil {
				log.Println(err)
				peer.Close()
				delete(sockets, peer)
			}
		}
	}
}
*/
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

// broadcastMsg to all peers
func broadcastMsg(h *Hub, m *Message) {
	b, err := json.Marshal(m)
	if err != nil {
		log.Println("broadcastMsg: msg failed to convert to []byte")
	}
	hub.broadcast <- b
}

// ConnectToPeers connects to the peers' addr:port by first parsing the string with the format
// addr1:port1, addr2:port2,...
func ConnectToPeers(newPeers string) {
	peers := strings.Split(newPeers, ",")
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

	var router *mux.Router
	InitHTTPServer(router, hub)
}
