package slackgw

import "math"

var eventNames = []string{
	"Invalid",
	"AccountsChangedEvent",
	"AckErrorEvent",
	"BotAddedEvent",
	"BotChangedEvent",
	"ChannelCreatedEvent",
	"ChannelHistoryChangedEvent",
	"ChannelInfoEvent",
	"ChannelJoinedEvent",
	"ChannelRenameEvent",
	"CommandsChangedEvent",
	"ConnectedEvent",
	"ConnectingEvent",
	"ConnectionErrorEvent",
	"DNDUpdatedEvent",
	"DisconnectedEvent",
	"EmailDomainChangedEvent",
	"EmojiChangedEvent",
	"FileCommentAddedEvent",
	"FileCommentDeletedEvent",
	"FileCommentEditedEvent",
	"GroupCreatedEvent",
	"GroupRenameEvent",
	"HelloEvent",
	"IMCreatedEvent",
	"InvalidAuthEvent",
	"ManualPresenceChangeEvent",
	"MessageEvent",
	"MessageTooLongEvent",
	"OutgoingErrorEvent",
	"PinAddedEvent",
	"PinRemovedEvent",
	"PrefChangeEvent",
	"PresenceChangeEvent",
	"ReactionAddedEvent",
	"ReactionRemovedEvent",
	"ReconnectUrlEvent",
	"StarAddedEvent",
	"StarRemovedEvent",
	"TeamDomainChangeEvent",
	"TeamJoinEvent",
	"TeamMigrationStartedEvent",
	"TeamPrefChangeEvent",
	"TeamRenameEvent",
	"UnmarshallingErrorEvent",
	"UserChangeEvent",
	"UserTypingEvent",
}

func EventNameToMask(name string) int64 {
	for i, n := range eventNames {
		if n == name {
			return int64(1<<uint64(i))
		}
	}
	return -1
}

func MaskToEventName(v int64) string {
	iv := int(math.Log2(float64(v)))
	if iv < 0 || iv >= len(eventNames) {
		iv = 0
	}
	return eventNames[iv]
}
