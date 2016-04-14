package slackgw

import "testing"

func TestEventString(t *testing.T) {
	for i := AccountsChangedEvent; i < MaxEvent; i = i << 1 {
		n := MaskToEventName(i)
		if n == "Invalid" {
			t.Errorf("Got invalid string for %d", i)
		}
		t.Logf("%s = %d", n, i)
	}

	if n := EventNameToMask("MessageEvent"); n != MessageEvent {
		t.Errorf("MessageEvent string should yield the correct mask, got %d", n)
	}
}