package slackgw

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"
)

func TestPostMessage(t *testing.T) {
	s0 := New()
	s := httptest.NewServer(s0)
	defer s.Close()

	t.Logf("Server started")

	// Wait for incoming message
	done := make(chan struct{})
	go func() {
		timeout := time.After(time.Second)
		select {
		case <-timeout:
			t.Errorf("timed out waiting for an event")
		case msg := <-s0.bus:
			t.Logf("%#v", msg)
			msg.dst <- nil
		}
		close(done)
	}()

	go http.PostForm(s.URL+"/post", url.Values{
		"channel": []string{"#test"},
		"message": []string{"Hello, World!"},
	})

	t.Logf("Waiting...")
	<-done
}