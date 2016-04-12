package slackgw

import "testing"

func TestEventString(t *testing.T) {
	for i := AccountsChangedEvent; i < MaxEvent; i = i << 1 {
		
t.Logf("%d", i)
		if MaskToEventName(i) == "Invalid" {
			t.Errorf("Got invalid string for %d", i)
		}
	}
}