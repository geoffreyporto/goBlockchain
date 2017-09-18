package main

import (
	"crypto/sha256"
	"encoding/binary"
	"log"
	"net/http"
	"reflect"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
)

// Block is the building blocks of a block chain
type Block struct {
	index     uint64
	timestamp int64
	prevHash  []byte
	data      []byte
	hash      []byte
}

const (
	MsgTypeQueryLatest    = 0
	MsgTypeQueryAll       = 1
	MsgTypeRespBlockchain = 2
)

const (
	HashLen = 32
	IntSize = 8
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
	return CalculateHash(block.index, block.timestamp, block.prevHash, block.data)
}

// GenerateNextBlock generates next block in the chain
func GenerateNextBlock(blockData []byte) *Block {
	prevBlock := GetLatestBlock()
	nextIndex := prevBlock.index + 1
	nextTimestamp := time.Now().UTC().Unix()
	nextHash := CalculateHash(nextIndex, nextTimestamp, prevBlock.prevHash, blockData)
	return &Block{nextIndex, nextTimestamp, prevBlock.prevHash, blockData, nextHash}
}

// GetGensisBlock generates and returns the genesis block
func GetGensisBlock() *Block {
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

// IsValidNewBlock checks the integrity of the newest block
func IsValidNewBlock(newBlock, prevBlock *Block) bool {
	if prevBlock.index+1 != newBlock.index {
		log.Println("IsValidNewBlock(): invalid index")
		return false
	} else if !reflect.DeepEqual(prevBlock.hash, newBlock.prevHash) {
		log.Println("IsValidNewBlock(): invalid previous hash")
		return false
	} else if !reflect.DeepEqual(CalculateHashForBlock(newBlock), newBlock.hash) {
		log.Println("invalid hash: %x is not %x", CalculateHashForBlock(newBlock), newBlock.hash)
		return false
	}
	return true
}

// IsValidChain checks if the chain received is valid
func IsValidChain(newBlockchain []Block) bool { // TODO: pass by ref?
	// newBlockchain is in JSON format
	// Check if the genesis block matches
	//genesisBlockRaw := []byte(newBlockchain[0])
	//var genesisBlock Block
	//err := json.Unmarshal(genesisBlockRaw, &genesisBlock)
	//if err != nil {
	//panic(err)
	//}
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
func ReplaceChain(newBlockchain []Block) { // TODO: pass by ref?
	if IsValidChain(newBlockchain) && (len(newBlockchain) > len(blockchain)) {
		log.Println("Received blockchain is valid. Replacing current blockchain with the received blockchain")
		blockchain = newBlockchain
		// broadcast(responseLatestMsg())
	} else {
		log.Println("Received blockchain invalid")
	}
}

// InitHTTPServer initializes the http server and REST api
// TODO: refactor http portion to a new file
func InitHTTPServer() {
	// need to pass these into command line arguments
	HttpPort := 3001

	// setup REST interface
	router = mux.NewRouter().StrictSlash(true)

	router.HandleFunc("/blocks", GetBlocks).Methods("GET")
	router.HandleFunc("/mineBlock", MineBlock).Methods("POST")
	router.HandleFunc("/peers", GetPeers).Methods("GET")
	router.HandleFunc("/addPeer", AddPeer).Methods("POST")

	log.Fatal(http.ListenAndServe(":"+strconv.Itoa(HttpPort), router))
}

// GetBlocks handles the /getBlocks REST request
func GetBlocks(w http.ResponseWriter, r *http.Request) {

}

// MineBlock handles the /mineBlock REST post
func MineBlock(w http.ResponseWriter, r *http.Request) {

}

// GetPeers handles the /peers REST request
func GetPeers(w http.ResponseWriter, r *http.Request) {

}

// AddPeer handles the /addPeer REST post
func AddPeer(w http.ResponseWriter, r *http.Request) {

}

// InitP2PServer sets up the websocket to other nodes
func InitP2PServer() {
	// need to pass these into command line arguments
	//P2pPort := 6001

}

var blockchain []Block
var router *mux.Router
var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

func main() {
	//var sockets

	// generate genesis block
	blockchain = append(blockchain, *GetGensisBlock())

}
