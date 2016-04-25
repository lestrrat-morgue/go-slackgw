package slackgw

import (
	"net/http"

	"github.com/nlopes/slack"
)

const (
	AccountsChangedEvent int64 = 1 << (iota + 1)
	AckErrorEvent
	BotAddedEvent
	BotChangedEvent
	ChannelCreatedEvent
	ChannelHistoryChangedEvent
	ChannelInfoEvent
	ChannelJoinedEvent
	ChannelRenameEvent
	CommandsChangedEvent
	ConnectedEvent
	ConnectingEvent
	ConnectionErrorEvent
	DNDUpdatedEvent
	DisconnectedEvent
	EmailDomainChangedEvent
	EmojiChangedEvent
	FileCommentAddedEvent
	FileCommentDeletedEvent
	FileCommentEditedEvent
	GroupCreatedEvent
	GroupRenameEvent
	HelloEvent
	IMCreatedEvent
	InvalidAuthEvent
	ManualPresenceChangeEvent
	MessageEvent
	MessageTooLongEvent
	OutgoingErrorEvent
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
	UnmarshallingErrorEvent
	UserChangeEvent
	UserTypingEvent
	MaxEvent
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
	AuthHeader string // if non empty, authorize
	AuthToken  string // XXX temporary. do not rely on this being here
	bus        chan *Message
	done       chan struct{}
	slack      SlackClient // For testing purposes, we use an interface here
	rtm        *slack.RTM
	rtmhandler SlackRTMHandler // Handles mesages
	slackuser  string
}
