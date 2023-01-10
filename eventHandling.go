package bombparty

import (
	"encoding/json"
	"fmt"
	"time"
)

func tryUnmarshal(b []byte, v any) {
	if err := json.Unmarshal(b, v); err != nil {
		panic("Couldn't unmarshal -> " + string(b) + " <- Error: " + err.Error())
	}
}

// Convert to Player & return a peer ID
func (r rawProfile) Player() (Player, int) {
	isLeader, isMod := false, false
	for _, role := range r.Roles {
		if role == "leader" {
			isLeader = true
		}
	}

	return Player{
		Username: r.Username,
		PFP:      r.PFP,
		Auth:     r.Auth,
		IsLeader: isLeader,
		IsMod:    isMod,
		ID:       fmt.Sprint(r.ID),
	}, r.ID
}

func (c *Client) eventHandle(data []byte) {
	out := []json.RawMessage{}

	if err := json.Unmarshal(data, &out); err != nil {
		panic("Unknown data type: " + string(data) + ": " + err.Error())
	}

	
	ev := ""
	if err := json.Unmarshal(out[0], &ev); err != nil {
		panic("Unknown data type: " + string(data) + ": " + err.Error())
	}
	
	switch ev {
	case "clearUsedWords":
		return
	}

	switch Event(ev) {
	case CHAT:
		// 42["chat",{"peerId":32,"auth":null,"roles":[],"nickname":"Guest7101"},"Guest8386 has returned"]
		authorRaw := rawProfile{}
		tryUnmarshal(out[1], &authorRaw)
		msg := ""
		tryUnmarshal(out[2], &msg)

		author := c.Room.Chatters[fmt.Sprint(authorRaw.ID)]
		sendEv(c, CHAT, &EventChat{
			Author:  author,
			Message: msg,
		})
	// TODO:
	case JOINED_GAME:
		resp := selfJoinResp{}
		err := json.Unmarshal(data, &resp)
		if err != nil {
			panic(err)
		}

		r := c.tmpRoom
		c.tmpRoom = nil

		r.Rules = RoomRules{
			Dictionary: resp.Rules.Dictionary.Value,
			// TODO:
			BonusAlphabet:    []string{},
			MinTurnDuration:  resp.Rules.MinTurnDuration.Value,
			PromptDifficulty: resp.Rules.PromptDifficulty.Value,
			MaxPromptAge:     resp.Rules.MaxPromptAge.Value,
			StartingLives:    resp.Rules.StartingLives.Value,
			MaxLives:         resp.Rules.MaxLives.Value,
		}

		for len(r.Chatters) == 0 {
		}
		r.Self = r.Chatters[fmt.Sprint(resp.SelfID)]

		if resp.Milestone.Name == "round" {
			r.Round = r.convertMilestone2Round(resp.Milestone)
		} else {
			if resp.Milestone.LastRound != nil {
				r.LastWinner = r.Chatters[fmt.Sprint(resp.Milestone.LastRound.Winner.ID)]
			}
		}
		for _, p := range resp.Players {
			id := fmt.Sprint(p.Profile.ID)
			r.PlayersWaiting[id] = r.Chatters[id]
		}

		c.Room = r
		if h, ok := c.eventMap[JOINED_GAME]; ok {
			h.(func(*EventJoinedGame))(&EventJoinedGame{
				Room: c.Room,
			})
		}
	case "setPlayerCount", "livesLost", "setRulesLocked":
		// ignore here

	case CHATTER_JOINED:
		c.Room.lastChatterID++
		chatter := rawProfile{}
		json.Unmarshal(data, &chatter)
		player, _ := chatter.Player()
		p := &player
		c.Room.Chatters[fmt.Sprint(c.Room.lastChatterID)] = p

		if h, ok := c.eventMap[CHATTER_JOINED]; ok {
			h.(func(*EventChatterJoined))(&EventChatterJoined{
				Player: p,
			})
		}
	case CHATTER_LEFT, "updatePlayer", "setSelfRoles", "setRoomPublic":
		// TODO:
		fmt.Println(string(ev), "->", string(data))
	case KICKED:
		reason := ""
		json.Unmarshal(data, &reason)

		if h, ok := c.eventMap[KICKED]; ok {
			h.(func(*EventSelfKicked))(&EventSelfKicked{
				Reason: reason,
			})
		}
	case PLAYER_JOINED_ROUND:
		rawData := rawAddPlayer{}
		json.Unmarshal(data, &rawData)
		id := fmt.Sprint(rawData.Profile.ID)
		p := c.Room.Chatters[id]

		c.Room.PlayersWaiting[id] = p

		if h, ok := c.eventMap[PLAYER_JOINED_ROUND]; ok {
			h.(func(*EventPlayerJoinedRound))(&EventPlayerJoinedRound{
				Player: p,
			})
		}
	case PLAYER_LEFT_ROUND:
		delete(c.Room.PlayersWaiting, string(data))
		if h, ok := c.eventMap[PLAYER_LEFT_ROUND]; ok {
			h.(func(*EventPlayerLeftRound))(&EventPlayerLeftRound{
				Player: c.Room.Chatters[string(data)],
			})
		}
	case COUNTDOWN:
		if h, ok := c.eventMap[COUNTDOWN]; ok {
			t := 0
			tRaw := []byte{}
			found := false
			for _, b := range data {
				if found {
					tRaw = append(tRaw, b)
				} else if b == 44 {
					found = true
				}
			}
			t = bToInt(tRaw)
			ti := time.UnixMilli(int64(t))
			h.(func(*EventCountdown))(&EventCountdown{
				ScheduledStart: ti,
			})
		}
	case "setMilestone":
		raw := milestone{}
		i := len(data) - 1
		for {
			if data[i] == 44 {
				data = data[:i]
				break
			}
			i--
		}
		json.Unmarshal(data, &raw)

		switch Event(raw.Name) {
		// Start round
		case ROUND_START:
			c.Room.Round = c.Room.convertMilestone2Round(raw)
			c.Room.PlayersWaiting = map[string]*Player{}

			if h, ok := c.eventMap[ROUND_START]; ok {
				h.(func(*EventRoundStart))(&EventRoundStart{
					TurnChange: &TurnChange{
						CurrentPlayer: c.Room.Round.Players[fmt.Sprint(raw.CurrentPlayer)],
						Prompt:        raw.Prompt,
					},
				})
			}
		case ROUND_END:
			c.Room.LastWinner = c.Room.Chatters[fmt.Sprint(raw.LastRound.Winner.ID)]
			winner := c.Room.Round.Players[fmt.Sprint(raw.LastRound.Winner.ID)]
			c.Room.Round = nil

			if h, ok := c.eventMap[ROUND_END]; ok {
				h.(func(*EventRoundEnd))(&EventRoundEnd{
					Winner: winner,
				})
			}
		default:
			panic("Unknown milestone '" + raw.Name + "'")
		}
		// Round statrt: 42["setMilestone",{"name":"round","startTime":1656771881309,"currentPlayerPeerId":1,"dictionaryManifest":{"name":"English","bonusAlphabet":"abcdefghijklmnopqrstuvwy","promptDifficulties":{"beginner":500,"medium":300,"hard":100}},"syllable":"cli","promptAge":0,"usedWordCount":0,"playerStatesByPeerId":{"0":{"lives":2,"word":"","wasWordValidated":false,"bonusLetters":[]},"1":{"lives":2,"word":"","wasWordValidated":false,"bonusLetters":[]}}},1656771881329]
		// Round end: 42["setMilestone",{"name":"seating","lastRound":{"winner":{"nickname":"Shady Goat","picture":"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"}},"rulesLocked":true,"dictionaryManifest":{"name":"English","bonusAlphabet":"abcdefghijklmnopqrstuvwy","promptDifficulties":{"beginner":500,"medium":300,"hard":100}}},1656771940331]
	case TURN_CHANGE:
		nextID := []byte{}
		nextPrompt := []byte{}
		readingPrompt := false
		for _, b := range data {
			if b == 44 {
				if readingPrompt {
					break
				} else {
					readingPrompt = true
				}
			} else if readingPrompt {
				if b == 34 {
					continue
				}
				nextPrompt = append(nextPrompt, b)
			} else {
				nextID = append(nextID, b)
			}
		}

		newPlayer := c.Room.Round.Players[string(nextID)]
		wasHurt := data[len(data)-1] == 49

		d := &TurnChange{
			LastPlayer:        c.Room.Round.CurrentPlayer,
			LastPlayerWasHurt: wasHurt,
			LastPrompt:        c.Room.Round.Prompt,
			CurrentPlayer:     newPlayer,
			Prompt:            string(nextPrompt),
		}
		if wasHurt {
			c.Room.Round.PromptAge++
			c.Room.Round.CurrentPlayer.Lives--
		}

		newPlayer.Guess = ""

		c.Room.Round.CurrentPlayer = newPlayer
		c.Room.Round.Prompt = string(nextPrompt)

		if h, ok := c.eventMap[TURN_CHANGE]; ok {
			h.(func(*EventTurnChange))(&EventTurnChange{
				TurnChange: d,
			})
		}
	case PLAYER_INPUT:
		id := []byte{}
		curInp := []byte{}
		wInput := false

		for _, b := range data {
			if b == 44 && !wInput {
				wInput = true
			} else if wInput {
				curInp = append(curInp, b)
			} else {
				id = append(id, b)
			}
		}
		curInp = curInp[1 : len(curInp)-1]
		player := c.Room.Round.Players[string(id)]
		player.Guess = string(curInp)

		if h, ok := c.eventMap[PLAYER_INPUT]; ok {
			h.(func(*EventPlayerInput))(&EventPlayerInput{
				RoundPlayer: player,
			})
		}
	case WORD_CORRECT:
		raw := rawWordCorrect{}
		json.Unmarshal(data, &raw)
		id := fmt.Sprint(raw.PlayerPeerID)
		p := c.Room.Round.Players[id]

		for _, l := range raw.BonusLetters {
			ru := []rune(l)[0]
			p.BonusLetters[ru] = true
		}

		if h, ok := c.eventMap[WORD_CORRECT]; ok {
			h.(func(*EventWordCorrect))(&EventWordCorrect{
				p,
			})
		}
	case "failWord":
		if h, ok := c.eventMap[WORD_FAIL]; ok {
			idRaw := []byte{}
			var r byte
			for n, b := range data {
				if b == 44 {
					r = data[n+2]
					break
				} else {
					idRaw = append(idRaw, b)
				}
			}

			e := &EventWordFail{
				Player: c.Room.Round.Players[string(idRaw)],
			}

			switch r {
			case 110:
				e.Reason = FR_NOT_IN_DICTIONARY
			case 109:
				e.Reason = FR_NO_PROMPT
			case 97:
				e.Reason = FR_ALREADY_USED
			}
			h.(func(*EventWordFail))(e)
		}
	case BONUS_ALPHABET_COMPLETED:
		idRaw := []byte{}
		health := []byte{}
		found := false

		for _, b := range data {
			if found {
				health = append(health, b)
			} else if b == 44 {
				found = true
			} else {
				idRaw = append(idRaw, b)
			}
		}
		hp := bToInt(health)

		p := c.Room.Round.Players[string(idRaw)]

		p.Lives = hp

		p.BonusLetters = map[rune]bool{}
		for _, l := range c.Room.Round.BonusAlphabet {
			p.BonusLetters[l] = false
		}

		if h, ok := c.eventMap[BONUS_ALPHABET_COMPLETED]; ok {
			h.(func(*EventBonusAlphabetComplete))(&EventBonusAlphabetComplete{
				Player: p,
			})
		}
	default:
		panic(string(ev) + " is unknown!")
	}
}

func sendEv[T any](c *Client, ev Event, v T) {
	if h, ok := c.eventMap[ev]; ok {
		h.(func(T))(v)
	}
}