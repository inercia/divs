package divsd

import (
	"testing"
)

func TestSwitchId(t *testing.T) {
	if testing.Short() {
		if NewSwitchId() == "" {
			t.Fail()
		}
	}
}
