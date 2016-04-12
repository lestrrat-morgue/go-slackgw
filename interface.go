package slackgw

import (
	"net/http"

	"github.com/nlopes/slack"
)

const (
	AccountsChangedEvent int32 = 1 << iota
	BotAddedEvent
	BotChangedEvent
	CommandsChangedEvent
	EmailDomainChangedEvent
	EmojiChangedEvent
	HelloEvent
	ManualPresenceChangeEvent
	MessageEvent
	PinAddedEvent
	PinRemovedEvent
	PrefChangeEvent
	PresenceChangeEvent
	ReactionAddedEvent
	ReactionRemovedEvent
	ReconnectUrlEvent
	StarAddedEvent
	StarRemovedEvent
	TeamDomainChangeEvent
	TeamJoinEvent
	TeamMigrationStartedEvent
	TeamPrefChangeEvent
	TeamRenameEvent
	UserChangeEvent
	UserTypingEvent
)

type SlackClient interface {
	NewRTM() *slack.RTM
	AuthTest() (*slack.AuthTestResponse, error)
	PostMessage(string, string, slack.PostMessageParameters) (string, string, error)
}

type SlackRTMClient interface {
	Disconnect() error
}

type Server struct {
	*http.ServeMux
	bus        chan *Message
	done       chan struct{}
	slack      SlackClient // For testing purposes, we use an interface here
	rtm        *slack.RTM
	rtmhandler SlackRTMHandler // Handles mesages
	slackuser  string
}
