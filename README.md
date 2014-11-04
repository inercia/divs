DiVS
=====

[![Go Walker](http://gowalker.org/api/v1/badge)](https://gowalker.org/github.com/inercia/divs)
[![Build Status](https://drone.io/github.com/inercia/divs/status.png)](https://drone.io/github.com/inercia/divs/latest)

## Overview

The DiVS is a distributed virtual switch. It allows you to create a switch where
the connected hosts are located at different networks but mutually reachable
through the Internet.

![Overview](https://raw.githubusercontent.com/inercia/divs/master/docs/images/overview.png)

In this example, machines A, B and C are located in different networks but
they are all connected to the Internet, so they can send traffic to each other.
In this scenario, DiVS creates a [TAP device](http://en.wikipedia.org/wiki/TUN/TAP)
(something like a regular ethernet device) in each of these machines, and then
it establishes something like a _P2P VPN_ between these three nodes where they appear
to be all connected to the same network segment. You must assign IP addresses to
these machines (the same you would do in a regular network). In this example, we
have assigned IPs in the 10.0.1.0/24 range, so the machines connected this way
with DiVS would have the illusion of being connected to a regular switch like
this:

![Equivalent Switch](https://raw.githubusercontent.com/inercia/divs/master/docs/images/equivalent-switch.png)

Terminology
-----------

  * `node`: each of the machines where a DiVS daemon is running (in this example,
  nodes A, B and C)
  * `tap device`: a device that simulates a link layer device and it operates
  with layer 2 packets like Ethernet frames. It can be used for creating a network bridge.
  * `endpoint`: a physical or virtual machine, with a MAC address, that is associated
  with a DiVS TAP device and sends and/or receives ethernet data.

Architecture
------------

DiVS uses the [memberlist](https://github.com/hashicorp/memberlist) library
for managing the virtual switch membership and member failure detection. `memberlist`
uses a gossip based protocol for spreading information (like nodes that are alive,
suspected to be down of definitely down), but `memberlist` also provides some useful
features like application-level messages, broadcasting of application-level
information, encryption, compression: we make use of these features for sending
virtual traffic between nodes.

DiVS maintains a eventually consistent _distributed database_ of MAC addresses,
mapping MAC addresses to nodes in the virtual network. This allows us to use the
TAP device for traffic to/from multiple endpoints in the same node (for example,
when using virtual machines in the physical machine). When a DiVS node `N` detects
a packet in the local TAP device with an unknown MAC address `M`, it updates the
distributed database setting `M -> N`. When other nodes whant to send traffic to
the MAC `M` they will encapsulate the packet and send it to the node `N`.

![MAC DiVS mapping](https://raw.githubusercontent.com/inercia/divs/master/docs/images/macs-table-overview.png)

The distributed database update is performed by using the gossip mechanism provided 
by `memberlist`. Nodes can add local information like these MAC to Node mappings
and `memberlist` will piggyback that information in the cluster management
messages. 

## Status

I'm currently going forward in the basic features of the distributed switch.
This is a bird's–eye view of the roadmap I have in mind:

+ [X] parse packets comming from the TAP device
+ [X] send the ethernet packets as UDP packets with the memberlist Send* facility.
+ [X] implement the distributed database for MAC addesses
+ [ ] return node-local ARP reponses, where nodes reqspond to ARP `who-is` queries in
      the local TAP device by using data from the distributed database
+ [ ] implement some kind of challenge-response in the initial connection between
      nodes (`memberlist` does not have anything like this, it just relies in
      encryption and both parties sharing the same key)
+ [ ] modify `memberlist` for being more NAT-friendly.
+ [ ] implement a DHCP server in the distributed switch. This would require either a) some
      consensus for not assigning the same IP to two different nodes or b) implementing
      IP pools per node 

## Installation

You can get several pre-build binaries from the [Drone.io](https://drone.io/github.com/inercia/divs/files)
continuous integration server. These builds cannot be considered stable (as they
are run after every push to the _github_ repository), but nothing here is very
stable yet...

If you prefer to build the software from the source code, you can checkout the
DiVS repository from _github_ with:

```sh
$ git clone https://github.com/inercia/divs.git
```

and then build the DiVS daemon with

```sh
$ make deps
$ make all
```

This will leave a binary, `divsd.exe`, at the top level directory. You can then
check out the command line arguments with:

```sh
$ ./divsd.exe --help
Usage: divsd.exe [global options] 
...
```

