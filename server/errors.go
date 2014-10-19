package divs

import (
	"fmt"
)

// Could not obtain a valid IP/port with NAT
var ERR_COULD_NOT_OBTAIN_NAT = fmt.Errorf("Could not obtain a valid IP/port")
