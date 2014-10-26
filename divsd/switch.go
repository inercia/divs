package divsd

import (
	"code.google.com/p/go-uuid/uuid"
)

// get a new switch device id
func NewSwitchId() string {
	// we could use base64 with base64.StdEncoding.EncodeToString(...)
	return uuid.NewUUID().String()
}
