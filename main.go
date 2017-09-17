package main

import (
	"crypto/sha256"
	"encoding/binary"
	"log"
	"net/http"
	"reflect"
	"time"
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
func IsValidChain(newBlockchain []Block) { // TODO: pass by ref?

}

// ReplaceChain checks and see if the current chain stored on the node needs to be replaced with
// a more up to date version (which is validated).
func ReplaceChain(newBlockchain []Block) { // TODO: pass by ref?
	if IsValidChain(newBlockchain) && (len(newBlockchain) > blockchain.length) {
		log.Println("Received blockchain is valid. Replacing current blockchain with the received blockchain")
		blockchain = newBlockchain
		// broadcast(responseLatestMsg())
	} else {
		log.Println("Received blockchain invalid")
	}
}

func InitHttpServer() {
	s := &http.Server {
		Addr: ":" + HttpPort,
		Handler:
	}
}

var blockchain []Block

func main() {
	//var sockets

	blockchain = append(blockchain, *GetGensisBlock())

}
