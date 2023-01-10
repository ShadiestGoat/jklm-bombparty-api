package bombparty

import "encoding/json"

func (r *Room) SendChat(msg string) {
	if r == nil || r.WS == nil {
		return
	}

	b, _ := json.Marshal(msg)

	wsBs := []byte(`42["chat",`)
	wsBs = append(wsBs, b...)
	wsBs = append(wsBs, ']')

	r.WS.WriteMessage(1, wsBs)
}
