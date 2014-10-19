package divs

import (
	"github.com/ccding/go-stun/stun"
	"github.com/huin/goupnp/dcps/internetgateway1"

	"fmt"
	"net"
	"strconv"
)

var ERR_COULD_NOT_OBTAIN_UPNP = fmt.Errorf("Could not obtain a valid IP/port with UPNP")

// get an external IP and port with UpnP
func Get() (ip string, port int, err error) {
	log.Debug("Using UPnP for getting external IP/port")

	clients, errors, err := internetgateway1.NewWANPPPConnection1Clients()
	if err != nil {
		log.Fatal(err)
	}

	log.Debug("Got %d errors finding UPnP servers. %d UPnP servers discovered.\n",
		len(errors), len(clients))
	for i, e := range errors {
		log.Error("Error finding server #%d: %v\n", i+1, e)
	}
	if len(clients) == 0 {
		return "", 0, ERR_COULD_NOT_OBTAIN_UPNP
	}

	for _, c := range clients {
		dev := &c.ServiceClient.RootDevice.Device
		srv := &c.ServiceClient.Service

		log.Debug(dev.FriendlyName, " :: ", srv.String())
		scpd, err := srv.RequestSCDP()
		if err != nil {
			log.Warning("  Error requesting service SCPD: %v\n", err)
		} else {
			log.Debug("  Available actions:")
			for _, action := range scpd.Actions {
				log.Debug("  * %s\n", action.Name)
				for _, arg := range action.Arguments {
					var varDesc string
					if stateVar := scpd.GetStateVariable(arg.RelatedStateVariable); stateVar != nil {
						varDesc = fmt.Sprintf(" (%s)", stateVar.DataType.Name)
					}
					log.Debug("    * [%s] %s%s\n", arg.Direction, arg.Name, varDesc)
				}
			}
		}

		if scpd == nil || scpd.GetAction("GetExternalIPAddress") != nil {
			ip, err := c.GetExternalIPAddress()
			log.Info("GetExternalIPAddress: ", ip, err)
		}
	}

	return ip, port, nil
}
