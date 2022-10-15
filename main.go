package main

import (
	"flag"
	"github.com/gorilla/websocket"
	"strings"
)

var dialer = websocket.Dialer{
	ReadBufferSize:  0,
	WriteBufferSize: 0,
}

func main() {
	usernanme := flag.String("u", "mother", "set username")
	roomRaw := flag.String("r", "", "room to connect to")
	flag.Parse()

	if *roomRaw == "" {
		panic("Room needed! Use -r!")
	}

	room := strings.ToUpper(*roomRaw)

	p := NewGuestClient(*usernanme)
	err := p.JoinRoom(room)
	if err != nil {
		panic(err)
	}
}
