package slackgw

import (
	"github.com/lestrrat/go-pdebug"
	"github.com/nlopes/slack"
)

type SlackRTMHandler interface {
	Handle(ctx *RTMCtx) error
}

type RTMCtx struct {
	UserID  string // This UserID is populated so handlers can potentially filter out messages addressed to others
	RTM     *slack.RTM
	Event   slack.RTMEvent
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
			if err := hdl.Handle(&RTMCtx{UserID: s.slackuser, RTM: rtm, Event: ev}); err != nil {
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
