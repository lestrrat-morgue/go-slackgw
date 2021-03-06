package main

import (
	"bytes"
	"flag"
	"fmt"
	"io/ioutil"
	"os"

	"golang.org/x/net/context"

	"google.golang.org/cloud/pubsub"

	"github.com/lestrrat/go-pdebug"
	"github.com/lestrrat/go-slackgw"
	"github.com/lestrrat/go-slackgw/gcp"
)

type eventList []int64

func (l eventList) String() string {
	buf := bytes.Buffer{}
	for i, v := range l {
		buf.WriteString(slackgw.MaskToEventName(v))
		if i != len(l)-1 {
			buf.WriteByte(',')
		}
	}
	return buf.String()
}

func (l *eventList) Set(v string) error {
	i := slackgw.EventNameToMask(v)
	if i == -1 {
		return fmt.Errorf("unknown event '%s'", v)
	}
	*l = append(*l, i)
	return nil
}

func main() {
	os.Exit(_main())
}

func _main() int {
	var listen string
	var token string
	var tokenf string
	var authtokenf string
	var projectID string
	var topic string
	var name string
	var rtm string
	var selfaddress bool
	var server bool
	var events eventList

	flag.StringVar(&listen, "listen", "127.0.0.1:4979", "listen address for HTTP interface")
	flag.StringVar(&token, "token", "", "Slack bot token")
	flag.StringVar(&tokenf, "tokenfile", "", "Slack bot token file")
	flag.StringVar(&authtokenf, "authtokenfile", "", "File containing token used to authentication when posting")
	flag.StringVar(&projectID, "gpubsub-forward.project_id", "", "Google Cloud Project ID")
	flag.StringVar(&topic, "gpubsub-forward.topic", "slackgw-forward", "topic to forward to")
	flag.Var(&events, "gpubsub-forward.event", "event(s) to forward")
	flag.BoolVar(&selfaddress, "gpubsub-forward.self-addressed-only", true, "forward only if it's address to this bot")
	flag.StringVar(&name, "name", "slackgw", "bot name")
	flag.StringVar(&rtm, "rtm", "", "RTM handler to enable (e.g. 'gpubsub-forward')")
	flag.BoolVar(&server, "server", true, "Turn on/off HTTP server")
	flag.Parse()

	s := slackgw.New()

	if token == "" {
		if tokenf == "" {
			token = os.Getenv("SLACK_API_TOKEN")
		} else {
			tokbuf, err := ioutil.ReadFile(tokenf)
			if err != nil {
				fmt.Printf("Failed to read from file %s: %s\n", tokenf, err)
				return 1
			}
			token = string(tokbuf)
		}
	}

	if token == "" {
		fmt.Printf("You must provide a Slack bot token\n")
		return 1
	}

	// Start slack client
	if err := s.StartSlack(token); err != nil {
		fmt.Printf("Failed to start slack client: %s\n", err)
		return 1
	}

	// Start HTTP Interface
	if server {
		if authtokenf != "" {
			buf, err := ioutil.ReadFile(authtokenf)
			if err != nil {
				fmt.Printf("Failed to read from '%s'", authtokenf)
				return 1
			}
			s.AuthToken = string(buf)
			s.AuthHeader = "X-Slackgw-Auth"
		}

		proto := "tcp" // hardcode for now
		if err := s.StartHTTP(proto, listen); err != nil {
			fmt.Printf("Failed to start HTTP server: %s\n", err)
			return 1
		}
	}

	// Enable RTM handler
	switch rtm {
	case "gpubsub-forward":
		hctx := context.Background()
		cl, err := pubsub.NewClient(hctx, projectID)
		if err != nil {
			fmt.Printf("Failed to create pubsub client: %s", err)
			return 1
		}

		fwd := gcp.NewPubsubForwarder(cl, topic, []int64(events)...)
		fwd.SelfAddressedOnly = selfaddress
		s.StartRTM(fwd)
	}

	// Wait till we're killed, or something goes wrong
	if err := s.Run(); err != nil {
		fmt.Printf("Failed to run: %s\n", err)
		return 1
	}

	if pdebug.Enabled {
		pdebug.Printf("Bailing out of '_main'\n")
	}

	return 0
}