package slackgw

import (
	"github.com/lestrrat/go-pdebug"
	"github.com/nlopes/slack"
)

type SlackRTMHandler interface {
	Handle(ctx *RTMCtx) error
}

type RTMCtx struct {
	RTM     *slack.RTM
	Event   slack.RTMEvent
	Message *slack.MessageEvent
}

func (s *Server) handleRTM() {
	if pdebug.Enabled {
		defer pdebug.Printf("Bailing out of handleRTM")
	}
	done := s.done
	rtm := s.rtm
	hdl := s.rtmhandler

	for loop := true; loop; {
		select {
		case ev := <-rtm.IncomingEvents:
			pdebug.Printf("Incoming!")
			if err := hdl.Handle(&RTMCtx{RTM: rtm, Event: ev}); err != nil {
				if pdebug.Enabled {
					pdebug.Printf("SlackRTMHandler: %s", err)
				}
				loop = false
			}
		case <-done:
			loop = false
		}
	}
}

// Reply replies to the channel where the message came from
func (ctx *RTMCtx) Reply(txt string) {
	ctx.RTM.SendMessage(ctx.RTM.NewOutgoingMessage(txt, ctx.Message.Channel))
}
