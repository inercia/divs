// The NAT package is responsible for
// * obtaining a pair of public IP and port that can be announced to external
//   nodes and can be used for sending traffic to this node.
// * keep that NAT traversal mechanism active, either by sending keepalives or
//   by notifying the corresponding service about our interest in keeping it active.
package nat

import (
	logging "github.com/op/go-logging"
)

const LOG_MODULE = "divs"

var log = logging.MustGetLogger(LOG_MODULE)
