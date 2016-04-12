package slackgw

import (
	"encoding/json"
	"errors"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"

	"github.com/lestrrat/go-pdebug"
	"github.com/nlopes/slack"
)

func New() *Server {
	mux := http.NewServeMux()
	s := &Server{ServeMux: mux}
	mux.HandleFunc("/", s.httpWelcome)
	mux.HandleFunc("/post", s.httpPostMessage)
	s.done = make(chan struct{})
	s.bus = make(chan *Message, 255)
	return s
}

func (s *Server) Close() error {
	// XXX lock?
	if s.done != nil {
		if pdebug.Enabled {
			pdebug.Printf("Closing 'done' channel...")
		}
		close(s.done)
		s.done = nil
	}

	if s.rtm != nil {
		if pdebug.Enabled {
			pdebug.Printf("Calling Disconnect() on RTM connection...")
		}
		s.rtm.Disconnect()
	}
	return nil
}

func (s *Server) StartHTTP(proto, listen string) error {
	if pdebug.Enabled {
		pdebug.Printf("Listening to %s:%s", proto, listen)
	}
	l, err := net.Listen(proto, listen)
	if err != nil {
		return err
	}
	// Start httpd
	go http.Serve(l, s)

	return nil
}

// StartSlack starts listening for incoming and outgoing events.
// Note that you will need to make sure to close the slack connection.
// In general you probably want to call StartSlack() and then Run()
func (s *Server) StartSlack(token string) error {
	cl := slack.New(token)
	s.slack = cl
	auth, err := s.slack.AuthTest()
	if err != nil {
		return err
	}
	s.slackuser = auth.UserID // so we know what to respond to

	// Start waiting for outgoing messages
	go s.watchOutgoingMessages()
	return nil
}

func (s *Server) StartRTM(h SlackRTMHandler) error {
	if pdebug.Enabled {
		pdebug.Printf("Starting RTM client...")
	}

	rtm := s.slack.NewRTM()
	s.rtm = rtm
	s.rtmhandler = h
	// Start listening to incoming messages
	go rtm.ManageConnection()

	// Start handling incoming message
	go s.handleRTM()
	return nil
}

func (s *Server) Run() error {
	// Setup a signal handler so we know when to properly disconnect
	sigch := make(chan os.Signal, 255)
	signal.Notify(sigch, syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)

	// Wait for close requests
	if pdebug.Enabled {
		pdebug.Printf("Looping...")
	}
	for loop := true; loop; {
		select {
		case <-s.done:
			if pdebug.Enabled {
				pdebug.Printf("Deteced 'done' state...")
			}
			loop = false
		case <-sigch:
			if pdebug.Enabled {
				pdebug.Printf("Received signal...")
			}
			s.Close()
			loop = false
		}
	}
	if pdebug.Enabled {
		pdebug.Printf("Bailing out of 'Run'")
	}
	return nil
}

func (s *Server) watchOutgoingMessages() {
	if pdebug.Enabled {
		pdebug.Printf("starting watchOutgoingMessages...")
		defer pdebug.Printf("Bailing out of watchOutgoingMessages")
	}

	done := s.done
	if done == nil {
		return
	}
	bus := s.bus
	client := s.slack

	for {
		select {
		case <-done:
			return
		case wrapped := <-bus:
			if pdebug.Enabled {
				pdebug.Printf("New outgoing message, sending to '%s'", wrapped.Channel)
			}
			_, _, err := client.PostMessage(wrapped.Channel, wrapped.Message, wrapped.Params)
			wrapped.dst <- err
		}
	}
}

func (s *Server) httpWelcome(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Welcome"))
}

func (s *Server) httpPostMessage(w http.ResponseWriter, r *http.Request) {
	if pdebug.Enabled {
		pdebug.Printf("http: posting new message...")
		defer pdebug.Printf("done posting new message")
	}

	msg, err := s.extractMessage(r)
	if err != nil {
		http.Error(w, "Failed to parse request: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer msgPool.Put(msg)

	// Note: this function WILL block until we get a response from the
	// slack server, because HTTP is... blocky.
	if err := s.postMessage(msg); err != nil {
		http.Error(w, "Failed to post message: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Sent"))
}

func (s *Server) postMessage(msg *Message) error {
	if s.done == nil {
		return errors.New("server is not connected or is shutting down")
	}

	dst := make(chan error)
	msg.dst = dst
	s.bus <- msg
	return <-dst
}

type Message struct {
	Channel string                      `json:"channel"`
	Message string                      `json:"message"`
	Params  slack.PostMessageParameters `json:"params"`
	dst     chan error                  // where we get the response
}

var msgPool = sync.Pool{New: allocMessage}

func allocMessage() interface{} {
	return &Message{Params: slack.NewPostMessageParameters()}
}

func releaseMessage(msg *Message) {
	msg.Channel = ""
	msg.Message = ""
	msg.Params = slack.NewPostMessageParameters()
	msg.dst = nil
	msgPool.Put(msg)
}

// extracts a usable slack.OutgoingMessage out of the request.
// TODO: allow takosan style messages to be parsed, too
func (s *Server) extractMessage(r *http.Request) (*Message, error) {
	var msg *Message
	switch m := r.Method; strings.ToLower(m) {
	case "post":
		switch ct := r.Header.Get("Content-Type"); ct {
		case "application/json":
			msg = msgPool.Get().(*Message)
			if err := json.NewDecoder(r.Body).Decode(msg); err != nil {
				return nil, err
			}
			return msg, nil
		case "application/x-www-form-urlencoded":
			msg = msgPool.Get().(*Message)
			msg.Channel = r.FormValue("channel")
			msg.Message = r.FormValue("message")
		default:
			return nil, errors.New("unknown content type: " + ct)
		}
	default:
		return nil, errors.New("unsupported method: " + m)
	}

	if msg.Channel == "" {
		defer msgPool.Put(msg)
		return nil, errors.New("channel cannot be empty")
	}

	if msg.Channel == "" {
		defer msgPool.Put(msg)
		return nil, errors.New("channel cannot be empty")
	}

	return msg, nil
}
