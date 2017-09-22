package main

import (
	"crypto/sha256"
	"encoding/binary"
	"encoding/json"
	"log"
	"net/http"
	"net/url"
	"os"
	"reflect"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
)

// Block is the building blocks of a block chain
type Block struct {
	Index     uint64 `json:"id"`
	Timestamp int64  `json:"timestamp"`
	PrevHash  []byte `json:"prevhash"`
	Data      []byte `json:"data"`
	Hash      []byte `json:"hash"`
}

// Message struct
type Message struct {
	Type uint    `json:"type"`
	Data []Block `json:"data"`
}

// Different message types for Message struct
const (
	MsgTypeQueryLatest    = 0
	MsgTypeQueryAll       = 1
	MsgTypeRespBlockchain = 2
)

// Other random constants
const (
	EnvPort     = "PORT"
	DefaultPort = "3001"
	HashLen     = 32
	IntSize     = 8
)

// CalculateHash calculates the hash of a block
func CalculateHash(index uint64, ts int64, prevHash []byte, data []byte) []byte {
	if len(prevHash) != HashLen {
		log.Fatal("CalculateHash(): Previous hash is not 32 bytes")
	}

	h := sha256.New()

	b := make([]byte, IntSize)
	binary.LittleEndian.PutUint64(b, index)
	h.Write(b)

	// uint64 conversion doesn't change the sign bit, only the way it's interpreted
	t := make([]byte, IntSize)
	binary.LittleEndian.PutUint64(t, uint64(ts))
	h.Write(t)

	h.Write(prevHash)
	h.Write(data)

	log.Printf("Calculated hash: %x\n", h.Sum(nil))
	return h.Sum(nil)
}

// CalculateHashForBlock calculates the has for the block
func CalculateHashForBlock(block *Block) []byte {
	return CalculateHash(block.Index, block.Timestamp, block.PrevHash, block.Data)
}

// GenerateNextBlock generates next block in the chain
func GenerateNextBlock(blockData []byte) *Block {
	prevBlock := GetLatestBlock()
	nextIndex := prevBlock.Index + 1
	nextTimestamp := time.Now().UTC().Unix()
	nextHash := CalculateHash(nextIndex, nextTimestamp, prevBlock.PrevHash, blockData)
	return &Block{nextIndex, nextTimestamp, prevBlock.PrevHash, blockData, nextHash}
}

// GetGenesisBlock generates and returns the genesis block
func GetGenesisBlock() *Block {
	var e = make([]byte, HashLen)
	timeNow := time.Now().UTC().Unix()
	genesisHash := CalculateHash(0, timeNow, e, e)
	return &Block{0, timeNow, e, e, genesisHash}
}

// GetLatestBlock retrieves the last block in the chain
func GetLatestBlock() *Block {
	// returns last element in slice
	return &blockchain[len(blockchain)-1]
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
	return &Message{MsgTypeRespBlockchain, blockchain}
}

// IsValidNewBlock checks the integrity of the newest block
func IsValidNewBlock(newBlock, prevBlock *Block) bool {
	if prevBlock.Index+1 != newBlock.Index {
		log.Println("IsValidNewBlock(): invalid index")
		return false
	} else if !reflect.DeepEqual(prevBlock.Hash, newBlock.PrevHash) {
		log.Println("IsValidNewBlock(): invalid previous hash")
		return false
	} else if !reflect.DeepEqual(CalculateHashForBlock(newBlock), newBlock.Hash) {
		log.Printf("invalid hash: %x is not %x\n", CalculateHashForBlock(newBlock), newBlock.Hash)
		return false
	}
	return true
}

// IsValidChain checks if the chain received is valid
func IsValidChain(newBlockchain []Block) bool { // TODO: pass by ref?
	// newBlockchain is in JSON format
	// Check if the genesis block matches
	if !reflect.DeepEqual(newBlockchain[0], blockchain[0]) {
		return false
	}

	var tempBlocks []Block
	tempBlocks = append(tempBlocks, newBlockchain[0])
	for i := 1; i < len(newBlockchain); i++ {
		if IsValidNewBlock(&newBlockchain[i], &tempBlocks[i-1]) {
			tempBlocks = append(tempBlocks, newBlockchain[i])
		} else {
			return false
		}
	}

	// All blocks are valid in the new blockchain
	return true
}

// ReplaceChain checks and see if the current chain stored on the node needs to be replaced with
// a more up to date version (which is validated).
func ReplaceChain(newBlockchain []Block) {
	if IsValidChain(newBlockchain) && (len(newBlockchain) > len(blockchain)) {
		log.Println("Received blockchain is valid. Replacing current blockchain with the received blockchain")
		blockchain = newBlockchain
		// broadcast(responseLatestMsg())
	} else {
		log.Println("Received blockchain invalid")
	}
}

// AddBlock validates and adds new block to blockchain
func AddBlock(newBlock Block) {
	if IsValidNewBlock(&newBlock, GetLatestBlock()) {
		blockchain = append(blockchain, newBlock)
	}
}

// InitHTTPServer initializes the http server and REST api
func InitHTTPServer() {
	// need to pass these into command line arguments
	// Get Env variable for HTTPPort
	port = os.Getenv(EnvPort)
	if port == "" {
		port = DefaultPort
	}

	// setup REST interface
	router = mux.NewRouter().StrictSlash(true)

	router.HandleFunc("/blocks", GetBlocks).Methods("GET")
	router.HandleFunc("/mineBlock", MineBlock).Methods("POST")
	router.HandleFunc("/peers", GetPeers).Methods("GET")
	router.HandleFunc("/addPeer", AddPeer).Methods("POST")
	// Starts websocket connections for peers
	router.HandleFunc("/ws", HandleWSConnection)

	log.Fatal(http.ListenAndServe(":"+port, router))
}

// GetBlocks handles the /blocks REST request
func GetBlocks(w http.ResponseWriter, r *http.Request) {
	// Send a copy of this node's blockchain
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(blockchain)
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
	AddBlock(block)
	//broadcast(responseLatestMsg())
}

// GetPeers handles the /peers REST request
func GetPeers(w http.ResponseWriter, r *http.Request) {
	// Sends list of peers this node is connected to
	log.Println("----Peers----")
	peersStr := ""
	for k, v := range sockets {
		if v == true {
			log.Println(k.RemoteAddr().String())
			peersStr += k.RemoteAddr().String()
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

// HandleWSConnection handles websocket connections to peers(other nodes)
func HandleWSConnection(w http.ResponseWriter, r *http.Request) {
	// Upgrade initial GET request to a websocket
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Fatal(err)
	}

	defer ws.Close()
	// Register our new client
	sockets[ws] = true

	for {
		var msg Message
		// Read in new message as JSON and map it to a Message object
		err := ws.ReadJSON(&msg)
		if err != nil {
			log.Println(err)
			delete(sockets, ws)
		}
		// Send the newly received message to the broadcast channel
		broadcast <- msg
	}
}

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

// InitConnection sets up the socket to pass messages
func InitConnection(c *websocket.Conn) {
	sockets[c] = true
	InitMessageHandler(c)
	InitErrorHandler(c)
	// TODO: Query chain length
}

// InitMessageHandler sets up the different types of messages it can receive

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
		InitConnection(c)
	}
}

var blockchain []Block
var port string
var router *mux.Router
var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

var sockets = make(map[*websocket.Conn]bool) // connected clients
var broadcast = make(chan Message)           // broadcast channel

func main() {
	//var sockets

	// generate genesis block
	var genesisBlock = *GetGenesisBlock()
	blockchain = append(blockchain, genesisBlock)

}
