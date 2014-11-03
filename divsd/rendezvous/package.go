// The rendezvous package is responsible for starting the discovery service
// The discovery service must:
// * announce in the network the presence if this node
// * lookup other nodes that are using the same service
package rendezvous

import (
	logging "github.com/op/go-logging"
)

const LOG_MODULE = "divs"

var log = logging.MustGetLogger(LOG_MODULE)
