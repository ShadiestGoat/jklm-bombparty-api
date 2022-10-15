package bombparty

import (
	"fmt"
	"math"
	"math/rand"
	"sort"
	"time"
)

// https://stackoverflow.com/questions/22892120/how-to-generate-a-random-string-of-a-fixed-length-in-go
const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

func randStringBytes(n int) string {
	rand.Seed(time.Now().UnixMilli())
	b := make([]byte, n)
	for i := range b {
		b[i] = letterBytes[rand.Intn(len(letterBytes))]
	}
	return string(b)
}

func bToInt(b []byte) int {
	n := 0
	cp := int(math.Pow10(len(b) - 1))
	for _, b := range b {
		n += int(b-48) * cp
		cp /= 10
	}
	return n
}

func (r Room) convertMilestone2Round(raw milestone) *Round {
	ba := []rune{}
	for _, ru := range raw.DictionaryManifest.BonusAlphabet {
		ba = append(ba, ru)
	}

	round := Round{
		Order:         []string{},
		CurrentPlayer: nil,
		Prompt:        raw.Prompt,
		Players:       map[string]*RoundPlayer{},
		BonusAlphabet: ba,
	}

	players := []string{}
	curID := fmt.Sprint(raw.CurrentPlayer)

	for id, state := range raw.PlayerStatesByPeerId {
		players = append(players, id)
		curBA := map[rune]bool{}

		for _, b := range ba {
			curBA[b] = false
		}
		for _, l := range state.BonusLetters {
			curBA[[]rune(l)[0]] = true
		}

		curP := &RoundPlayer{
			Player:       r.Chatters[id],
			Lives:        state.Lives,
			Guess:        state.Input,
			BonusLetters: curBA,
		}

		round.Players[id] = curP
		if id == curID {
			round.CurrentPlayer = curP
		}
		if id == r.Self.ID {
			round.Self = curP
		}
	}

	sort.Slice(players, func(i, j int) bool {
		return players[i] > players[j]
	})

	round.Order = players

	return &round
}
