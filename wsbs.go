package main

import (
	"encoding/json"
	"fmt"
)

func (c *Client) wsMain(r *Room) {
	conn, _, _ := dialer.Dial("wss://"+r.Server+".jklm.fun/socket.io/?EIO=4&transport=websocket", nil)
	r.WS = conn

	defer r.WS.Close()

	for {
		_, msgRaw, err := r.WS.ReadMessage()
		// TODO: Error handle!
		if err != nil {
			r.WS.Close()
			r.WS = nil
			break
		}

		operation, rest := extractOperation(msgRaw)
		switch operation {
		case ev_hello:
			r.WS.WriteMessage(1, []byte(ev_hello_authed))
		case ev_ping:
			r.WS.WriteMessage(1, []byte(ev_pong))
		case ev_hello_resp:
			r.WS.WriteMessage(1, []byte(`42["joinGame","bombparty","`+r.Code+`","`+c.Token+`"]`))
			c.tmpRoom = r
		default:
			// fmt.Println(string(operation), string(rest))
			c.eventHandle(rest)
		}
	}
}

func (c *Client) wsAuth(r *Room) {
	defer r.WSAuth.Close()

	for {
		_, msgRaw, err := r.WSAuth.ReadMessage()
		// TODO: Actually error handle!
		if err != nil {
			r.WSAuth.Close()
			r.WSAuth = nil
			break
		}

		operation, rest := extractOperation(msgRaw)
		switch operation {
		case ev_ping:
			r.WSAuth.WriteMessage(1, []byte(ev_pong))
		case ev_hello:
			r.WSAuth.WriteMessage(1, []byte(ev_hello_resp))
		case ev_hello_resp:
			r.WSAuth.WriteMessage(1, []byte(`420["joinRoom",{"roomCode":"`+r.Code+`","userToken":"`+c.Token+`","nickname":"`+c.Username+`","language":"en-US"}]`))
		case ev_room_joined:
			go c.wsMain(r)
			r.WSAuth.WriteMessage(1, []byte(`421["getChatterProfiles"]`))
		case ev_chatter_profile:
			rest = rest[1 : len(rest)-1]
			profiles := []rawProfile{}
			json.Unmarshal(rest, &profiles)
			for _, prof := range profiles {
				player, id := prof.Player()
				r.lastChatterID = id
				r.Chatters[fmt.Sprint(id)] = &player
			}
		case ev_gameEvents:
			c.eventHandle(rest)
		}
	}
}

func (c *Client) Close() {
	if c.Room != nil {
		c.Room.WS.Close()
		c.Room.WS = nil
		c.Room.WSAuth.Close()
		c.Room.WSAuth = nil
	}
}

func extractOperation(inp []byte) (Event, []byte) {
	operation := []byte{}
	breakLoc := 0

	for i, b := range inp {
		if 48 <= b && b <= 57 {
			operation = append(operation, b)
		} else {
			breakLoc = i
			break
		}
	}

	return Event(operation), inp[breakLoc:]
}
