package divsd

import (
	"fmt"
)

// Could not obtain a valid IP/port with NAT
var ERR_COULD_NOT_OBTAIN_NAT = fmt.Errorf("Could not obtain a valid IP/port")

// Could not parse address
var ERR_COULD_NOT_PARSE_ADDR = fmt.Errorf("Could not parse address")

// Malformed message
var ERR_MALFORMED_MSG = fmt.Errorf("Malformed message")
