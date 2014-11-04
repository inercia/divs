package divsd

import (
	"testing"
)

func TestSwitchId(t *testing.T) {
	if testing.Short() {
		sw := NewSwitchId()
		if !sw.Empty() {
			t.Fail()
		}
	}
}
