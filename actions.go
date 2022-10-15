package bombparty

import "fmt"

func (c *Client) SetGuess(guess string, submit bool) {
	c.Room.WS.WriteMessage(1, []byte(fmt.Sprintf(`42["setWord","%v",%v]`, guess, submit)))
}
