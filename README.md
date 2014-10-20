DiVS
=====

[![Go Walker](http://gowalker.org/api/v1/badge)](https://gowalker.org/github.com/inercia/divs)

## Overview

The DiVS server is a distributed virtual switch.

It allows you to create a switch where the connected hosts are located at different networks but mutually reacheable through the Internet.

![Overview](https://raw.githubusercontent.com/inercia/divs/master/docs/images/overview.png)

It uses the [goraft](https://github.com/goraft/raft) library.

## Running

First, install DiVS with:

```sh
$ git clone https://github.com/inercia/divs.git
```

