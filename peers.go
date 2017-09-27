package main

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

const (
	// Time allowed to write a message to the peer.
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer.
	pongWait = 60 * time.Second

	// Send pings to peer with this period. Must be less than pongWait.
	pingPeriod = (pongWait * 9) / 10

	// Maximum message size allowed from peer.
	maxMessageSize = 512
)

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
			log.Println("Registering: ", peer)
		case peer := <-h.unregister:
			log.Println("Unregister: ", peer)
			if _, ok := h.peers[peer]; ok {
				delete(h.peers, peer)
				close(peer.send)
			}
		case message := <-h.broadcast:
			log.Println("Broadcasting: ", message)
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

// TODO: use this vs using p.conn.WriteJSON()
func (p *Peer) sendMsg(m *Message) {
	encodedMsg, err := json.Marshal(*m)
	//encodedMsg, err := binary.Write(m)
	if err != nil {
		log.Println("sendMsg: msg failed to convert to []byte")
	}
	p.send <- encodedMsg
}

// broadcastMsg to all peers
func (h *Hub) broadcastMsg(m *Message) {
	b, err := json.Marshal(*m)
	if err != nil {
		log.Println("broadcastMsg: msg failed to convert to []byte")
	}
	log.Println("brodcastMsg: ", m)
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
	/*
		if err := peer.conn.WriteJSON(QueryChainLengthMsg()); err != nil {
			log.Println("addWS: failed to query longest chain")
		}*/

	// Allow collection of memory referenced by the caller by doing all work in
	// new goroutines.
	go peer.writePump()
	go peer.readPump()

	peer.sendMsg(QueryChainLengthMsg())
}

// readPump pumps messages from the websocket connection to the hub.
//
// The application runs readPump in a per-connection goroutine. The application
// ensures that there is at most one reader on a connection by executing all
// reads from this goroutine.
func (p *Peer) readPump() {
	defer func() {
		p.hub.unregister <- p
		p.conn.Close()
	}()
	p.conn.SetReadLimit(maxMessageSize)
	p.conn.SetReadDeadline(time.Now().Add(pongWait))
	p.conn.SetPongHandler(func(string) error { p.conn.SetReadDeadline(time.Now().Add(pongWait)); return nil })

	for {
		var receivedMsg Message
		err := p.conn.ReadJSON(&receivedMsg)
		if err != nil {
			log.Printf("error: %v", err)
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway) {
				log.Printf("unexpected close error: %v", err)
			}
			break
		}
		// Do the writing in the writePump()
		switch receivedMsg.Type {
		case MsgTypeQueryAll:
			//write response chain message
			//p.conn.WriteJSON(QueryChainLengthMsg())
			p.sendMsg(QueryChainLengthMsg())
		case MsgTypeQueryLatest:
			// write repsonse latest msg
			//p.conn.WriteJSON(QueryAllMsg())
			p.sendMsg(QueryAllMsg())
		case MsgTypeRespBlockchain:
			receivedMsg.HandleBlockchainResp()
		default:
			log.Fatal("Unknown message type")
		}
	}
}

// writePump pumps messages from the hub to the websocket connection.
//
// A goroutine running writePump is started for each connection. The
// application ensures that there is at most one writer to a connection by
// executing all writes from this goroutine.
func (p *Peer) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		p.conn.Close()
	}()

	for {
		select {
		case message, ok := <-p.send:
			p.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				// The hub closed the channel.
				p.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			err := p.conn.WriteMessage(websocket.TextMessage, message)
			if err != nil {
				return
			}
		case <-ticker.C:
			p.conn.SetWriteDeadline(time.Now().Add(writeWait))
			err := p.conn.WriteMessage(websocket.PingMessage, []byte{})
			if err != nil {
				return
			}
		}
	}
}
