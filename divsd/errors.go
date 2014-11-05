package divsd

import (
	"fmt"
)

// Could not parse address
var ERR_COULD_NOT_PARSE_ADDR = fmt.Errorf("Could not parse address")

// Malformed message
var ERR_MALFORMED_MSG = fmt.Errorf("Malformed message")

// Timeout while waiting for peers
var ERR_TIMEOUT_PEERS = fmt.Errorf("Timeout while waiting for peers")
