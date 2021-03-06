package gcp

import (
	"bytes"
	"encoding/gob"
	"encoding/json"
	"errors"
	"regexp"
	"strings"
	"sync"
	"time"

	"golang.org/x/net/context"

	"github.com/lestrrat/go-pdebug"
	"github.com/lestrrat/go-slackgw"
	"github.com/nlopes/slack"

	"google.golang.org/cloud/pubsub"
)

// EventForwarder creates a new slackgw.SlackRTMHandler that forwards the
// specified events
type PubsubForwarder struct {
	initonce          sync.Once
	mask              int64 // 25 events
	client            *pubsub.Client
	pubch             chan slack.RTMEvent
	topic             string
	SelfAddressedOnly bool // only accept messages that are address to this bot
}

func init() {
	gob.Register(slack.MessageEvent{})
}

//	hctx := context.Background()
//	cl, err := pubsub.NewClient(hctx, projectID)
//	if err != nil {
//		return err
//	}
//	NewPubsubForwarder(cl, ....)
func NewPubsubForwarder(cl *pubsub.Client, topic string, events ...int64) *PubsubForwarder {
	var mask int64
	for _, ev := range events {
		mask |= ev
	}

	return &PubsubForwarder{
		mask:   mask,
		pubch:  make(chan slack.RTMEvent),
		client: cl,
		topic:  topic,
	}
}

type SlackLink struct {
	Text string
	URL  string
}

func parseSlackLink(s string) (*SlackLink, error) {
	if len(s) == 0 || s[0] != '<' {
		return nil, errors.New("not a link")
	}
	sl := &SlackLink{}
	for i := 1; i < len(s); i++ {
		switch s[i] {
		case '|':
			sl.Text = s[1:i]
		case '>':
			if l := len(sl.Text); l > 0 {
				sl.URL = sl.Text
				sl.Text = s[len(sl.Text)+2 : i]
			} else {
				sl.Text = s[1:i]
			}
			return sl, nil
		}
	}

	return nil, errors.New("not a link")
}

var rxSplitWS = regexp.MustCompile(`\s+`)

func (f *PubsubForwarder) Handle(ctx *slackgw.RTMCtx) error {
	ev := ctx.Event
	mask := f.mask

	f.initonce.Do(func() {
		go f.loop()
	})

	if pdebug.Enabled {
		pdebug.Printf("New event: %#v", ev)
	}

	switch ev.Data.(type) {
	case *slack.AccountsChangedEvent:
		if (mask & slackgw.AccountsChangedEvent) == 0 {
			return nil
		}
	case *slack.AckErrorEvent:
		if (mask & slackgw.AckErrorEvent) == 0 {
			return nil
		}
	case *slack.BotAddedEvent:
		if (mask & slackgw.BotAddedEvent) == 0 {
			return nil
		}
	case *slack.BotChangedEvent:
		if (mask & slackgw.BotChangedEvent) == 0 {
			return nil
		}
	case *slack.ChannelCreatedEvent:
		if (mask & slackgw.ChannelCreatedEvent) == 0 {
			return nil
		}
	case *slack.ChannelHistoryChangedEvent:
		if (mask & slackgw.ChannelHistoryChangedEvent) == 0 {
			return nil
		}
	case *slack.ChannelInfoEvent:
		if (mask & slackgw.ChannelInfoEvent) == 0 {
			return nil
		}
	case *slack.ChannelJoinedEvent:
		if (mask & slackgw.ChannelJoinedEvent) == 0 {
			return nil
		}
	case *slack.ChannelRenameEvent:
		if (mask & slackgw.ChannelRenameEvent) == 0 {
			return nil
		}
	case *slack.CommandsChangedEvent:
		if (mask & slackgw.CommandsChangedEvent) == 0 {
			return nil
		}
	case *slack.ConnectedEvent:
		if (mask & slackgw.ConnectedEvent) == 0 {
			return nil
		}
	case *slack.ConnectingEvent:
		if (mask & slackgw.ConnectingEvent) == 0 {
			return nil
		}
	case *slack.ConnectionErrorEvent:
		if (mask & slackgw.ConnectionErrorEvent) == 0 {
			return nil
		}
	case *slack.DNDUpdatedEvent:
		if (mask & slackgw.DNDUpdatedEvent) == 0 {
			return nil
		}
	case *slack.DisconnectedEvent:
		if (mask & slackgw.DisconnectedEvent) == 0 {
			return nil
		}
	case *slack.EmailDomainChangedEvent:
		if (mask & slackgw.EmailDomainChangedEvent) == 0 {
			return nil
		}
	case *slack.EmojiChangedEvent:
		if (mask & slackgw.EmojiChangedEvent) == 0 {
			return nil
		}
	case *slack.FileCommentAddedEvent:
		if (mask & slackgw.FileCommentAddedEvent) == 0 {
			return nil
		}
	case *slack.FileCommentDeletedEvent:
		if (mask & slackgw.FileCommentDeletedEvent) == 0 {
			return nil
		}
	case *slack.FileCommentEditedEvent:
		if (mask & slackgw.FileCommentEditedEvent) == 0 {
			return nil
		}
	case *slack.GroupCreatedEvent:
		if (mask & slackgw.GroupCreatedEvent) == 0 {
			return nil
		}
	case *slack.GroupRenameEvent:
		if (mask & slackgw.GroupRenameEvent) == 0 {
			return nil
		}
	case *slack.HelloEvent:
		if (mask & slackgw.HelloEvent) == 0 {
			return nil
		}
	case *slack.IMCreatedEvent:
		if (mask & slackgw.IMCreatedEvent) == 0 {
			return nil
		}
	case *slack.InvalidAuthEvent:
		if (mask & slackgw.InvalidAuthEvent) == 0 {
			return nil
		}
	case *slack.ManualPresenceChangeEvent:
		if (mask & slackgw.ManualPresenceChangeEvent) == 0 {
			return nil
		}
	case *slack.MessageEvent:
		if (mask & slackgw.MessageEvent) == 0 {
			return nil
		}

		if f.SelfAddressedOnly {
			d := ev.Data.(*slack.MessageEvent)
			// Parse the first word, and make sure it's addressed to uskk
			words := rxSplitWS.Split(strings.TrimSpace(d.Text), 2)
			if len(words) <= 0 {
				return nil
			}

			l, err := parseSlackLink(words[0])
			if err != nil {
				return nil
			}
			if l.Text != "@"+ctx.UserID {
				return nil
			}
		}
	case *slack.MessageTooLongEvent:
		if (mask & slackgw.MessageTooLongEvent) == 0 {
			return nil
		}
	case *slack.OutgoingErrorEvent:
		if (mask & slackgw.OutgoingErrorEvent) == 0 {
			return nil
		}
	case *slack.PinAddedEvent:
		if (mask & slackgw.PinAddedEvent) == 0 {
			return nil
		}
	case *slack.PinRemovedEvent:
		if (mask & slackgw.PinRemovedEvent) == 0 {
			return nil
		}
	case *slack.PrefChangeEvent:
		if (mask & slackgw.PrefChangeEvent) == 0 {
			return nil
		}
	case *slack.PresenceChangeEvent:
		if (mask & slackgw.PresenceChangeEvent) == 0 {
			return nil
		}
	case *slack.ReactionAddedEvent:
		if (mask & slackgw.ReactionAddedEvent) == 0 {
			return nil
		}
	case *slack.ReactionRemovedEvent:
		if (mask & slackgw.ReactionRemovedEvent) == 0 {
			return nil
		}
	case *slack.ReconnectUrlEvent:
		if (mask & slackgw.ReconnectUrlEvent) == 0 {
			return nil
		}
	case *slack.StarAddedEvent:
		if (mask & slackgw.StarAddedEvent) == 0 {
			return nil
		}
	case *slack.StarRemovedEvent:
		if (mask & slackgw.StarRemovedEvent) == 0 {
			return nil
		}
	case *slack.TeamDomainChangeEvent:
		if (mask & slackgw.TeamDomainChangeEvent) == 0 {
			return nil
		}
	case *slack.TeamJoinEvent:
		if (mask & slackgw.TeamJoinEvent) == 0 {
			return nil
		}
	case *slack.TeamMigrationStartedEvent:
		if (mask & slackgw.TeamMigrationStartedEvent) == 0 {
			return nil
		}
	case *slack.TeamPrefChangeEvent:
		if (mask & slackgw.TeamPrefChangeEvent) == 0 {
			return nil
		}
	case *slack.TeamRenameEvent:
		if (mask & slackgw.TeamRenameEvent) == 0 {
			return nil
		}
	case *slack.UnmarshallingErrorEvent:
		if (mask & slackgw.UnmarshallingErrorEvent) == 0 {
			return nil
		}
	case *slack.UserChangeEvent:
		if (mask & slackgw.UserChangeEvent) == 0 {
			return nil
		}
	case *slack.UserTypingEvent:
		if (mask & slackgw.UserTypingEvent) == 0 {
			return nil
		}
	default:
		return nil
	}
	f.pubch <- ev

	return nil
}

func (f *PubsubForwarder) loop() {
	if pdebug.Enabled {
		pdebug.Printf("Start gcp.PubsubForwarder.loop()")
		defer pdebug.Printf("Bailing out of gcp.PubsubForwarder.loop()")
	}

	flusht := time.Tick(time.Second)
	topic := f.client.Topic(f.topic)
	buf := make([]slack.RTMEvent, 0, pubsub.MaxPublishBatchSize)
	msgs := make([]*pubsub.Message, 0, pubsub.MaxPublishBatchSize)
	for {
		select {
		case ev := <-f.pubch:
			buf = append(buf, ev)
			if len(buf) <= pubsub.MaxPublishBatchSize {
				continue
			}
		case <-flusht:
			if len(buf) == 0 {
				continue
			}
		}

		if pdebug.Enabled {
			pdebug.Printf("Processing %d events...", len(buf))
		}

		jsbuf := bytes.Buffer{}
		enc := json.NewEncoder(&jsbuf)
		for _, ev := range buf {
			jsbuf.Reset()
			if err := enc.Encode(ev); err != nil {
				if pdebug.Enabled {
					pdebug.Printf("ERROR: %s", err)
				}
				// Ugh. Ignore
				continue
			}
			msgs = append(msgs, &pubsub.Message{Data: jsbuf.Bytes()})
		}
		buf = buf[:0]

		// TODO: handle errors
		if pdebug.Enabled {
			pdebug.Printf("Forwarding %d messages to %s", len(msgs), topic.Name())
		}

		res, err := topic.Publish(context.Background(), msgs...)
		if pdebug.Enabled {
			if err != nil {
				pdebug.Printf("%s", err)
			}
			if res != nil {
				pdebug.Printf("%#v", res)
			}
		}
		msgs = msgs[:0]
	}
}