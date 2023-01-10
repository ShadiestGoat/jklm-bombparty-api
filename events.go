package bombparty

import (
	"time"

	"github.com/gorilla/websocket"
)

// All the events that could happen
type Event string

// TODO: Change these from numbers to stuff like JoinGame!!!
const (
	ev_hello           Event = "0"
	ev_hello_resp      Event = "40"
	ev_hello_authed    Event = "40"
	ev_ping            Event = "2"
	ev_pong            Event = "3"
	ev_join_room       Event = "420"
	ev_room_joined     Event = "430"
	ev_chatter_profile Event = "431"
	ev_gameEvents      Event = "42"

	JOINED_GAME              Event = "setup"
	PLAYER_JOINED_ROUND      Event = "addPlayer"
	PLAYER_LEFT_ROUND        Event = "removePlayer"
	CHATTER_JOINED           Event = "chatterAdded"
	CHATTER_LEFT             Event = "chatterRemoved"
	COUNTDOWN                Event = "setStartTime"
	ROUND_START              Event = "round"
	ROUND_END                Event = "seating"
	TURN_CHANGE              Event = "nextTurn"
	PLAYER_INPUT             Event = "setPlayerWord"
	WORD_CORRECT             Event = "correctWord"
	WORD_FAIL                Event = "failWord"
	BONUS_ALPHABET_COMPLETED Event = "bonusAlphabetCompleted"
	KICKED                   Event = "kicked"
)

// The reason why a guess failed.
// Used as an enum with constants starting with FR_
type FailReason int

const (
	// The guess made isn't a recognised word
	FR_NOT_IN_DICTIONARY FailReason = iota
	// The guess didn't contain the prompt
	FR_NO_PROMPT
	// The guess has been used already
	FR_ALREADY_USED
)

/*
>====================<
       Raw JSON
>====================<

This is the raw json response.
These should not be public to avoid confusion.
*/

type selfJoinResp struct {
	SelfID   int `json:"selfPeerId"`
	LeaderID int `json:"leaderPeerId"`
	Rules    struct {
		Dictionary       rule[string] `json:"dictionaryId"`
		MinTurnDuration  rule[int]    `json:"minTurnDuration"`
		PromptDifficulty rule[int]    `json:"customPromptDifficulty"`
		MaxPromptAge     rule[int]    `json:"maxPromptAge"`
		StartingLives    rule[int]    `json:"startingLives"`
		MaxLives         rule[int]    `json:"maxLives"`
	} `json:"rules"`
	Players   []rawAddPlayer `json:"players"`
	Milestone milestone      `json:"milestone"`
}

type rule[T any] struct {
	Value T `json:"value"`
}

type playerState struct {
	Lives        int      `json:"lives"`
	Input        string   `json:"word"`
	WasValidated bool     `json:"wasWordValidated"`
	BonusLetters []string `json:"bonusLetters"`
}

type lastRound struct {
	Winner rawProfile `json:"winner"`
}

type rawProfile struct {
	ID       int         `json:"peerId"`
	Username string      `json:"nickname"`
	PFP      string      `json:"picture"`
	Auth     *PlayerAuth `json:"auth"`
	Roles    []string    `json:"roles"`
}

type rawAddPlayer struct {
	Profile rawProfile `json:"profile"`
}

// "dictionaryManifest":{"name":"English","bonusAlphabet":"abcdefghijklmnopqrstuvwy","promptDifficulties":{"beginner":500,"medium":300,"hard":100}}

type milestone struct {
	Name                 string                 `json:"name"`
	CurrentPlayer        int                    `json:"currentPlayerPeerId"`
	Prompt               string                 `json:"syllable"`
	PlayerStatesByPeerId map[string]playerState `json:"playerStatesByPeerId"`
	LastRound            *lastRound             `json:"lastRound"`
	DictionaryManifest   struct {
		BonusAlphabet string `json:"bonusAlphabet"`
	} `json:"dictionaryManifest"`
}

type roomCodeResp struct {
	Server string `json:"url"`
}

type PlayerAuth struct {
	Service  string `json:"service"`
	Username string `json:"username"`
	ID       string `json:"id"`
}

type rawWordCorrect struct {
	PlayerPeerID int      `json:"playerPeerId"`
	BonusLetters []string `json:"bonusLetters"`
}

/*
>====================<
      Components
>====================<

These are things that events will wrap
*/

// Rules for this room
type RoomRules struct {
	// The name of the dictionary used
	Dictionary string
	// An array of letters needed to recover a life
	BonusAlphabet []string
	// The minimum turn duration, in seconds
	MinTurnDuration int
	// If PromptDifficulty < 0, then this is at most X words per prompt.
	// Otherwise, this is at least X words per prompt.
	PromptDifficulty int
	// If this many players fail, the prompt changes.
	MaxPromptAge int
	// Starting amount of lives
	StartingLives int
	// Maximum lives
	MaxLives int
}

type Round struct {
	Order         []string
	CurrentPlayer *RoundPlayer
	Self          *RoundPlayer
	Prompt        string
	Players       map[string]*RoundPlayer
	PromptAge     int
	BonusAlphabet []rune
}

// Room (refered to as 'milestone' by the API) describes the room the player is in.
type Room struct {
	// A WebSocket connection to a room
	WS     *websocket.Conn
	WSAuth *websocket.Conn

	// The room code. An all-caps 4 letter string.
	Code string

	// Room rules
	Rules RoomRules
	// If Round == nil, there is no active round in this room.
	Round *Round
	// LastWinner will be nil if a a roung is being played, or if there is no previous round
	LastWinner *Player

	// The prefix that is used for a connection to this room.
	// I believe it's used for load balancing.
	Server string
	// This player's ID.
	Self *Player
	// All the connected chatters. This should be used as a repository of available players.
	Chatters map[string]*Player
	// Players waiting for the next round. To add your own player, see Client.JoinRound().
	// If a round is happening, this will be an empty map. This should not be used for verifying if a round is happening.
	PlayersWaiting map[string]*Player
	// used for finding which ID a new chatter will be assigned
	lastChatterID int
}

type RoundPlayer struct {
	*Player
	Lives int
	Guess string
	// This will always be false when it is this player's turn.
	// The tells the client after their turn is over if they ended up with a correct guess or not.
	GuessWasRight bool
	BonusLetters  map[rune]bool
}

type Player struct {
	Username string
	PFP      string
	Auth     *PlayerAuth
	IsLeader bool
	IsMod    bool
	ID       string
}

type TurnChange struct {
	LastPlayer        *RoundPlayer
	LastPlayerWasHurt bool
	LastPrompt        string

	CurrentPlayer *RoundPlayer
	Prompt        string
}

/*
>====================<
        Events
>====================<

These are all events that users can add for handlers.
They should all start with Event* (eg. EventJoinedGame).
*/

type EventJoinedGame struct {
	*Room
}

type EventChatterJoined struct {
	*Player
}

type EventPlayerJoinedRound struct {
	*Player
}

type EventSelfKicked struct {
	Reason string
}

type EventPlayerLeftRound struct {
	*Player
}

// TODO: Idr what the fuck this is???
type BeginCountdown struct {
}

type EventRoundStart struct {
	*TurnChange
}

type EventRoundEnd struct {
	Winner *RoundPlayer
}

// Used whenever a player looses lives
type EventLivesLost struct {
	Player     *RoundPlayer
	AmountLost int
}

// Whenever the turn changes
type EventTurnChange struct {
	*TurnChange
}

// Whenever a player updates their guess
type EventPlayerInput struct {
	*RoundPlayer
}

type EventWordCorrect struct {
	*RoundPlayer
}

type EventWordFail struct {
	Player *RoundPlayer
	Reason FailReason
}

// Start the 15 second countdown
type EventCountdown struct {
	ScheduledStart time.Time
}

type EventBonusAlphabetComplete struct {
	Player *RoundPlayer
}

func (c *Client) AddEventHandler(handler interface{}) {
	switch handler.(type) {
	case func(*EventJoinedGame):
		c.eventMap[JOINED_GAME] = handler
	case func(*EventChatterJoined):
		c.eventMap[CHATTER_JOINED] = handler
	case func(*EventPlayerJoinedRound):
		c.eventMap[PLAYER_JOINED_ROUND] = handler
	case func(*EventPlayerLeftRound):
		c.eventMap[PLAYER_LEFT_ROUND] = handler
	case func(*EventRoundStart):
		c.eventMap[ROUND_START] = handler
	case func(*EventRoundEnd):
		c.eventMap[ROUND_END] = handler
	case func(*EventWordCorrect):
		c.eventMap[WORD_CORRECT] = handler
	case func(*EventWordFail):
		c.eventMap[WORD_FAIL] = handler
	case func(*EventTurnChange):
		c.eventMap[TURN_CHANGE] = handler
	case func(*EventPlayerInput):
		c.eventMap[PLAYER_INPUT] = handler
	case func(*EventSelfKicked):
		c.eventMap[KICKED] = handler
	}
}

func (c *Client) RemoveEventHandler(event Event) {
	delete(c.eventMap, event)
}
