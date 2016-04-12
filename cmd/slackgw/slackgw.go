package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"

	"golang.org/x/net/context"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/pubsub/v1"

	"github.com/lestrrat/go-pdebug"
	"github.com/lestrrat/go-slackgw"
	"github.com/lestrrat/go-slackgw/gcp"
)

func main() {
	os.Exit(_main())
}

func _main() int {
	var listen string
	var token string
	var tokenf string
	var topic string
	var name string
	var icon string
	var rtm string
	var server bool

	flag.StringVar(&listen, "listen", "127.0.0.1:4979", "listen address for HTTP interface")
	flag.StringVar(&token, "token", "", "Slack bot token")
	flag.StringVar(&tokenf, "tokenfile", "", "Slack bot token file")
	flag.StringVar(&topic, "topic", "projects/:project_id:/topic/slackgw-forward", "topic to forward to")
	flag.StringVar(&name, "name", "slackgw", "bot name")
	flag.StringVar(&icon, "icon", "https://raw.githubusercontent.com/kentaro/slackgw/master/slackgw.jpg", "icon for slackgw")
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
		httpcl, err := google.DefaultClient(hctx, pubsub.PubsubScope)
		if err != nil {
			fmt.Printf("Failed to create default oauth client: %s\n", err)
			return 1
		}
		pubsubsvc, err := pubsub.New(httpcl)
		if err != nil {
			fmt.Printf("Failed to create pubsub client: %s\n", err)
			return 1
		}
		s.StartRTM(gcp.NewPubsubForwarder(pubsubsvc, topic, slackgw.MessageEvent))
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