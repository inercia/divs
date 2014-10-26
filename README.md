DiVS
=====

[![Go Walker](http://gowalker.org/api/v1/badge)](https://gowalker.org/github.com/inercia/divs)
[![Build Status](https://drone.io/github.com/inercia/divs/status.png)](https://drone.io/github.com/inercia/divs/latest)

## Overview

The DiVS server is a distributed virtual switch.

It allows you to create a switch where the connected hosts are located at
different networks but mutually reachable through the Internet.

![Overview](https://raw.githubusercontent.com/inercia/divs/master/docs/images/overview.png)

DiVS mantains a distributed database of MAC addresses that associate each MAC
to a node in the virtual topology. This feature wouldn't be strictly necessary (as
ethernet is not a reliable medium), but I plan to make use of the
[goraft](https://github.com/goraft/raft) library for the distributed database, so
it should be easy to add some reliability and consistency to the MAC to Node
mapping...

## Status

I'm currently going forward in the basic features of the distributed switch.

+ [ ] implement the distributed database for MAC addesses
+ [ ] parse packets comming from the TAP device
+ [ ] send the ethernet packets as UDP packets with the memberlist Send* facility.

## Running

First, install DiVS with:

```sh
$ git clone https://github.com/inercia/divs.git
```

