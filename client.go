package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

const BASE_URL = "https://jklm.fun/api"

type Client struct {
	// Username of the client
	Username string
	// Token used to identify the client
	Token string

	eventMap map[Event]interface{}
	Room     *Room
	tmpRoom  *Room
}

func NewGuestClient(username string) *Client {
	return &Client{
		Username: username,
		Token:    randStringBytes(16),
		eventMap: map[Event]interface{}{},
	}
}

func (c *Client) JoinRoom(Code string) error {
	reqBody := bytes.NewReader([]byte(`{"roomCode":"` + Code + `"}`))
	req, _ := http.NewRequest("POST", BASE_URL+"/joinRoom", reqBody)
	req.Header.Add("content-type", "application/json")
	req.Header.Add("origin", "https://jklm.fun")
	req.Header.Add("referer", "https://jklm.fun/"+Code)

	respRaw, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	respBody, _ := io.ReadAll(respRaw.Body)
	if (string(respBody)) == `{"errorCode":"noSuchRoom"}` {
		return ErrNoRoom
	}
	serv := roomCodeResp{}
	json.Unmarshal(respBody, &serv)
	server := strings.SplitN(serv.Server, ".", 2)[0]
	server = server[8:]

	connAuth, _, err := dialer.Dial("wss://"+server+".jklm.fun/socket.io/?EIO=4&transport=websocket", nil)

	if err != nil {
		return err
	}

	room := &Room{
		WSAuth:         connAuth,
		Code:           Code,
		Server:         server,
		Chatters:       map[string]*Player{},
		PlayersWaiting: map[string]*Player{},
	}

	go c.wsAuth(room)

	return nil
}

func (c *Client) JoinRound() error {
	if c.Room == nil {
		return ErrNotConnected
	}
	c.Room.WS.WriteMessage(1, []byte(`42["joinRound"]`))
	return nil
}

// TODO:
func (c *Client) LeaveRound() error {
	return nil
}

// TODO:
func (c *Client) LeaveRoom() error {
	return nil
}

// TODO:
func (c *Client) MakeRoom(name string, public bool) error {
	reqBody := strings.NewReader(fmt.Sprintf(`{"name":"%v","isPublic":%v,"gameId":"bombparty","creatorUserToken":"%v"`, name, public, c.Token))
	req, _ := http.NewRequest("POST", BASE_URL+"/joinRoom", reqBody)
	req.Header.Add("content-type", "application/json")

	// TODO:
	_, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}

	return nil
}
