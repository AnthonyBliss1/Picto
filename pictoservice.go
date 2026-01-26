package main

import (
	"encoding/json"
	"fmt"
	"log"
	"log/slog"
	"net"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/hashicorp/mdns"
)

type RoomChoice string

var LanIP net.IP

const (
	RoomChoiceCreate RoomChoice = "create_room"
	RoomChoiceJoin   RoomChoice = "join_room"
)

type Room struct {
	HostName string `json:"hostname"`
	Addr     string `json:"addr"`
	Port     int    `json:"port"`
	URL      string `json:"url"`
}

type Picto struct {
	RoomChoice     *RoomChoice  `json:"roomChoice"`
	AvailableRooms []Room       `json:"availableRooms"`
	CurrentRoom    *Room        `json:"currentRoom"`
	MDNSServer     *mdns.Server `json:"mdnsServer"` // ? might remove
	WsServer       *http.Server `json:"sServer"`
	Hub            *Hub         `json:"hub"`
	Mu             sync.Mutex
}

func init() {
	var err error

	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	slog.SetDefault(logger)

	LanIP, err = LANIPv4()
	if err != nil {
		log.Fatalf("Could not get LAN IP: %f", err)
	}
}

func (p *Picto) String() string {
	j, _ := json.MarshalIndent(p, "", "  ")
	return string(j)
}

// Handling Rooms
// ~~~~~~~~~~~~~~~~~~~

func (p *Picto) GetAvailableRooms() []Room {
	p.Mu.Lock()
	defer p.Mu.Unlock()

	// return a copy so caller can’t mutate slice
	out := make([]Room, len(p.AvailableRooms))
	copy(out, p.AvailableRooms)
	return out
}

func (p *Picto) GetCurrentroom() *Room {
	return p.CurrentRoom
}

func (p *Picto) SetCurrentRoom(r *Room, isUserHost bool) (ok bool, err error) {
	if isUserHost {

		hostName, err := os.Hostname()
		if err != nil {
			slog.Error("Failed to get user hostName", "Error", err)
			return false, err
		}

		url := fmt.Sprintf("ws://%s:%d/ws", LanIP.String(), 8000)

		userHost := Room{HostName: hostName, Addr: LanIP.String(), Port: 8000, URL: url}
		p.CurrentRoom = &userHost
		p.Hub = NewHub() // assign a hub for the host

	} else {
		p.CurrentRoom = r
	}

	return true, nil
}

func (p *Picto) IsHost() bool {
	switch {
	case p.Hub != nil:
		return true

	case p.Hub == nil:
		return false

	default:
		return false
	}
}

// Handling MDNS Server - MDNS Advertising Port: 8000 -> WS Port: 8000
// ~~~~~~~~~~~~~~~~~~~~~~~~~

func (p *Picto) MDNSLookup() error {
	entriesCH := make(chan *mdns.ServiceEntry, 4)
	newRooms := []Room{}

	go func() {
		for entry := range entriesCH {
			if !strings.Contains(entry.Name, "_pictosvelte._tcp") {
				continue
			}

			if entry.Port != 8000 {
				continue
			}

			slog.Debug("MDNS Lookup", "Found new entry", entry.Name)

			room := Room{HostName: entry.Host, Addr: entry.AddrV4.String(), Port: entry.Port}
			room.URL = fmt.Sprintf("ws://%s:%d/ws", room.Addr, room.Port)
			newRooms = append(newRooms, room)
		}
		p.Mu.Lock()
		p.AvailableRooms = newRooms
		p.Mu.Unlock()
	}()

	if err := mdns.Lookup("_pictosvelte._tcp", entriesCH); err != nil {
		return err
	}

	close(entriesCH)

	return nil
}

func (p *Picto) StartServers() (ok bool, err error) {
	hostName, _ := os.Hostname()

	info := []string{"Picto Server"}
	service, _ := mdns.NewMDNSService(hostName, "_pictosvelte._tcp", "", "", 8000, []net.IP{LanIP}, info)

	slog.Debug("Starting MDNS Server advertising service on :8000")
	mdnsSrv, err := mdns.NewServer(&mdns.Config{Zone: service})
	if err != nil {
		return false, err
	}

	p.MDNSServer = mdnsSrv

	p.StartWsServer()

	return true, nil
}

// Handling WS Server - Client, Hub, and Messages
// ~~~~~~~~~~~~~~~~~~~~~~~~~

type Message struct {
	Action      string  `json:"action"`
	Phase       string  `json:"phase"`
	Points      []Point `json:"points"`
	StrokeWidth int     `json:"strokeWidth"`
	Color       string  `json:"color"`
}

type Point struct {
	X float32 `json:"x"`
	Y float32 `json:"y"`
}

func (p *Picto) PrintMessage(msg *Message) string {
	b, _ := json.MarshalIndent(msg, "", "  ")
	return string(b)
}

type Client struct {
	hub  *Hub
	conn *websocket.Conn
	send chan Message
}

type Hub struct {
	Clients    map[*Client]bool
	Broadcast  chan Message
	Register   chan *Client
	Unregister chan *Client
}

func NewHub() *Hub {
	return &Hub{
		Clients:    make(map[*Client]bool),
		Broadcast:  make(chan Message),
		Register:   make(chan *Client),
		Unregister: make(chan *Client),
	}
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.Register:
			h.Clients[client] = true

		case client := <-h.Unregister:
			if _, ok := h.Clients[client]; ok {
				delete(h.Clients, client)
				close(client.send)
			}

		case msg := <-h.Broadcast:
			for client := range h.Clients {
				select {
				case client.send <- msg:
				default:
					close(client.send)
					delete(h.Clients, client)
				}
			}
		}
	}
}

func (p *Picto) StartWsServer() {
	if p.Hub == nil {
		return
	}

	go p.Hub.Run()

	mux := http.NewServeMux()
	mux.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		ServeWs(p.Hub, w, r)
	})

	p.WsServer = &http.Server{
		Addr:    LanIP.String() + ":8000",
		Handler: mux,
	}

	slog.Debug("Starting WebSocket Server on :8000")
	go func() {
		if err := p.WsServer.ListenAndServe(); err != nil {
			slog.Error("Server shutdown error", "Error", err)
		}
	}()
}

func (p *Picto) StopServers() {
	if p.WsServer != nil {
		p.WsServer.Close()
		p.MDNSServer.Shutdown()
	}
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

func ServeWs(hub *Hub, w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		slog.Error("Error upgrading Http to WS", "Error", err)
		return
	}
	client := &Client{hub: hub, conn: conn, send: make(chan Message, 256)}
	client.hub.Register <- client

	go client.writePump()
	go client.readPump()
}

// Handling Read and Writes to WS
// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

const (
	writeWait  = 10 * time.Second
	pongWait   = 60 * time.Second
	pingPeriod = (pongWait * 9) / 10
)

func (c *Client) readPump() {
	defer func() {
		c.hub.Unregister <- c
		c.conn.Close()
	}()

	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error { c.conn.SetReadDeadline(time.Now().Add(pongWait)); return nil })

	for {
		var msg *Message

		err := c.conn.ReadJSON(&msg)
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				slog.Error("WebSocket Error", "Error", err)
			}
			break
		}
		c.hub.Broadcast <- *msg
	}
}

func (c *Client) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()
	for {
		select {
		case msg, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				return
			}

			c.conn.WriteJSON(&msg)

		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}
