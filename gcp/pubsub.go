package gcp

import (
	"bytes"
	"encoding/base64"
	"encoding/gob"
	"encoding/json"
	"sync"
	"time"

	"github.com/lestrrat/go-pdebug"
	"github.com/lestrrat/go-slackgw"
	"github.com/nlopes/slack"

	"google.golang.org/api/pubsub/v1"
)

// EventForwarder creates a new slackgw.SlackRTMHandler that forwards the
// specified events
type PubsubForwarder struct {
	initonce sync.Once
	mask     int32 // 25 events
	pubch    chan slack.RTMEvent
	svc      *pubsub.Service
	topic    string
}

func init() {
	gob.Register(slack.MessageEvent{})
}

//	hctx := context.Background()
//	httpcl, err := google.DefaultClient(hctx, pubsub.PubsubScope)
//	if err != nil {
//		return err
//	}
//	pubsubsvc, err := pubsub.New(httpcl)
//	if err != nil {
//		return err
//	}
//	NewPubsubForwarder(pubsubsvc, ....)
func NewPubsubForwarder(svc *pubsub.Service, topic string, events ...int32) slackgw.SlackRTMHandler {
	var mask int32
	for _, ev := range events {
		mask |= ev
	}

	pdebug.Printf("mask = %d", mask)
	return &PubsubForwarder{
		mask:  mask,
		pubch: make(chan slack.RTMEvent),
		svc:   svc,
		topic: topic,
	}
}

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
	case *slack.BotAddedEvent:
		if (mask & slackgw.BotAddedEvent) == 0 {
			return nil
		}
	case *slack.BotChangedEvent:
		if (mask & slackgw.BotChangedEvent) == 0 {
			return nil
		}
	case *slack.CommandsChangedEvent:
		if (mask & slackgw.CommandsChangedEvent) == 0 {
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
	case *slack.HelloEvent:
		if (mask & slackgw.HelloEvent) == 0 {
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
	default:
		return nil
	}
	/*
	*slack.PinAddedEvent
	*slack.PinRemovedEvent
	*slack.PrefChangeEvent
	*slack.PresenceChangeEvent
	*slack.ReactionAddedEvent
	*slack.ReactionRemovedEvent
	*slack.ReconnectUrlEvent
	*slack.StarAddedEvent
	*slack.StarRemovedEvent
	*slack.TeamDomainChangeEvent
	*slack.TeamJoinEvent
	*slack.TeamMigrationStartedEvent
	*slack.TeamPrefChangeEvent
	*slack.TeamRenameEvent
	*slack.UserChangeEvent
	*slack.UserTypingEvent
	 */
	f.pubch <- ev

	return nil
}

func (f *PubsubForwarder) loop() {
	svc := f.svc
	topic := f.topic
	buf := make([]slack.RTMEvent, 0, 255)
	msgs := make([]*pubsub.PubsubMessage, 0, 255)
	flusht := time.Tick(2 * time.Second)
	b64enc := base64.RawURLEncoding
	for {
		select {
		case ev := <-f.pubch:
			buf = append(buf, ev)
			if len(buf) <= 255 {
				continue
			}
		case <-flusht:
			if len(buf) == 0 {
				continue
			}
		}

		encbuf := bytes.Buffer{}
		enc := json.NewEncoder(&encbuf)
		for _, ev := range buf {
			encbuf.Reset()
			if err := enc.Encode(ev); err != nil {
				pdebug.Printf("ERROR: %s", err)
				// Ugh. Ignore
				continue
			}
			msgs = append(msgs, &pubsub.PubsubMessage{Data: b64enc.EncodeToString(encbuf.Bytes())})
		}
		buf = buf[:0]

		pdebug.Printf("msgs = %#v", msgs)

		// TODO: handle errors
		res, err := svc.Projects.Topics.Publish(topic, &pubsub.PublishRequest{Messages: msgs}).Do()
		if err != nil {
			pdebug.Printf("%s", err)
		}
		pdebug.Printf("%#v", res)
		msgs = msgs[:0]
	}
}