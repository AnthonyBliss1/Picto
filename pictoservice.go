package main

import (
	"encoding/json"
	"fmt"
	"log/slog"
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
	HostName string
	Addr     string
	Port     int
	URL      string
}

type Picto struct {
	RoomChoice     *RoomChoice
	AvailableRooms []Room

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

func (p *Picto) GetAvailableRooms() []Room {
	p.Mu.Lock()
	defer p.Mu.Unlock()

	// return a copy so caller can’t mutate slice
	out := make([]Room, len(p.AvailableRooms))
	copy(out, p.AvailableRooms)
	return out
}

func (p *Picto) MDNSLookup() {
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

			slog.Info("Found new entry:", entry)

			room := Room{HostName: entry.Host, Addr: entry.AddrV4.String(), Port: entry.Port}
			room.URL = fmt.Sprintf("ws://%s:%d/ws", room.Addr, room.Port)
			newRooms = append(newRooms, room)
		}
		p.Mu.Lock()
		p.AvailableRooms = newRooms
		p.Mu.Unlock()
	}()

	mdns.Lookup("_pictosvelte._tcp", entriesCH)
	close(entriesCH)
}
