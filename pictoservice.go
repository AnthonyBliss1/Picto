package main

import (
	"encoding/json"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os"
	"strings"
	"sync"

	"github.com/hashicorp/mdns"
)

type RoomChoice string

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
	IsHost         bool         `json:"isHost"`
	MDNSServer     *mdns.Server `json:"mdnsServer"`

	Mu sync.Mutex
}

func init() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	slog.SetDefault(logger)
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

func (p *Picto) SetCurrentRoom(r *Room, isUserHost bool) error {
	if isUserHost {

		hostName, err := os.Hostname()
		if err != nil {
			slog.Error("Failed to get user hostName", err)
			return err
		}

		url := fmt.Sprintf("ws://%s:%d/ws", "0.0.0.0", 8000)

		userHost := Room{HostName: hostName, Addr: "0.0.0.0", Port: 8000, URL: url}
		p.CurrentRoom = &userHost

	} else {
		p.CurrentRoom = r
	}

	return nil
}

func (p *Picto) GetIsHost() bool {
	return p.IsHost
}

func (p *Picto) SetIsHost(b bool) {
	p.IsHost = b
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

func (p *Picto) StartServers() error {
	hostName, _ := os.Hostname()

	info := []string{"Picto Server"}
	service, _ := mdns.NewMDNSService(hostName, "_pictosvelte._tcp", "", "", 8000, nil, info)

	slog.Debug("Starting MDNS Server advertising service on :8000")
	mdnsSrv, err := mdns.NewServer(&mdns.Config{Zone: service})
	if err != nil {
		return err
	}

	p.MDNSServer = mdnsSrv

	go func() {
		p.StartWsServer()
	}()

	return nil
}

// Handling WS Server
// ~~~~~~~~~~~~~~~~~~~~~~~~~

func (p *Picto) StartWsServer() {
	mux := http.NewServeMux()
	// mux.HandleFunc("/ws", p.HandleConnections)

	server := &http.Server{
		Addr:    "0.0.0.0:8000",
		Handler: mux,
	}

	slog.Debug("Starting WebSocket Server on :8000")
	if err := server.ListenAndServe(); err != nil {
		log.Printf("server shutdown error: %v\n", err)
	}
}
