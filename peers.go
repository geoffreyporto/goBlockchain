package main

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

var (
	newline = []byte{'\n'}
	space   = []byte{' '}
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

// Peer is a middleman between the websocket connection and the hub.
type Peer struct {
	hub *Hub

	// The websocket connection.
	conn *websocket.Conn

	// Buffered channel of outbound messages.
	send chan []byte
}

// Hub maintains the set of active peers and broadcasts messages to the peers.
type Hub struct {
	// Registered clients.
	peers map[*Peer]bool

	// Inbound messages from the clients.
	broadcast chan []byte

	// Register requests from the clients.
	register chan *Peer

	// Unregister requests from clients.
	unregister chan *Peer
}

func newHub() *Hub {
	return &Hub{
		broadcast:  make(chan []byte),
		register:   make(chan *Peer),
		unregister: make(chan *Peer),
		peers:      make(map[*Peer]bool),
	}
}

func (h *Hub) run() {
	for {
		select {
		case peer := <-h.register:
			h.peers[peer] = true
		case peer := <-h.unregister:
			if _, ok := h.peers[peer]; ok {
				delete(h.peers, peer)
				close(peer.send)
			}
		case message := <-h.broadcast:
			for peer := range h.peers {
				select {
				case peer.send <- message:
				default:
					close(peer.send)
					delete(h.peers, peer)
				}
			}
		}
	}
}

func (p *Peer) sendMsg(m *Message) {
	encodedMsg, err := json.Marshal(m)
	if err != nil {
		log.Println("sendMsg: msg failed to convert to []byte")
	}
	p.send <- encodedMsg
}

// broadcastMsg to all peers
func (h *Hub) broadcastMsg(m *Message) {
	b, err := json.Marshal(m)
	if err != nil {
		log.Println("broadcastMsg: msg failed to convert to []byte")
	}
	hub.broadcast <- b
}

// serveWs handles websocket requests from the peer.
func serveWs(hub *Hub, w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}
	addWs(hub, conn)
}

// addWs adds new websocket connection to the hub
func addWs(hub *Hub, conn *websocket.Conn) {
	peer := &Peer{hub: hub, conn: conn, send: make(chan []byte, 256)}
	peer.hub.register <- peer

	// query for the longest chain
	peer.sendMsg(QueryChainLengthMsg())

	// Allow collection of memory referenced by the caller by doing all work in
	// new goroutines.
	//go peer.writePump()
	//go peer.readPump()
}
