package divsd

import (
	"encoding/base64"
	"encoding/hex"

	"code.google.com/p/go-uuid/uuid"
)

type UUID struct {
	uuid.UUID
}

// Get a new switch device id
func NewSwitchId() UUID {
	// we could use base64 with base64.StdEncoding.EncodeToString(...)
	return UUID{UUID: uuid.NewUUID()}
}

// Get a new switch device id from a string
func NewSwitchFromString(s string) UUID {
	return UUID{UUID: uuid.Parse(s)}
}

// Get a new UUID in Base64
func (uuid UUID) ToBase64() string {
	return base64.StdEncoding.EncodeToString([]byte(uuid.UUID))
}

// Get a new UUID in hexadecimal
func (uuid UUID) ToHex() string {
	return hex.EncodeToString([]byte(uuid.UUID))
}

// Return `true` if the UUID is empty
func (uuid UUID) Empty() bool {
	return len(uuid.UUID) == 0
}
