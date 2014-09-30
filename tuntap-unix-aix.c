/*
 * Copyright (c) 2012 Tristan Le Guern <leguern AT medu DOT se>
 *
 * Permission to use, copy, modify, and distribute this software for any
 * purpose with or without fee is hereby granted, provided that the above
 * copyright notice and this permission notice appear in all copies.
 *
 * THE SOFTWARE IS PROVIDED "AS IS" AND THE AUTHOR DISCLAIMS ALL WARRANTIES
 * WITH REGARD TO THIS SOFTWARE INCLUDING ALL IMPLIED WARRANTIES OF
 * MERCHANTABILITY AND FITNESS. IN NO EVENT SHALL THE AUTHOR BE LIABLE FOR
 * ANY SPECIAL, DIRECT, INDIRECT, OR CONSEQUENTIAL DAMAGES OR ANY DAMAGES
 * WHATSOEVER RESULTING FROM LOSS OF USE, DATA OR PROFITS, WHETHER IN AN
 * ACTION OF CONTRACT, NEGLIGENCE OR OTHER TORTIOUS ACTION, ARISING OUT OF
 * OR IN CONNECTION WITH THE USE OR PERFORMANCE OF THIS SOFTWARE.
 */

#include <sys/types.h>
#include <sys/ioctl.h>
#include <sys/socket.h>

#include <arpa/inet.h>
#include <net/if.h>
#include <net/if_tun.h>
#include <net/if_types.h>
#include <netinet/if_ether.h>
#include <netinet/in.h>

#include <fcntl.h>
#include <stdint.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <unistd.h>

#include "tuntap.h"

static int
tuntap_sys_create_dev(struct device *dev, int tun) {
	return -1;
}

int
tuntap_sys_start(struct device *dev, int mode, int tun) {
	return -1;
}

void
tuntap_sys_destroy(struct device *dev) {
	return;
}

int
tuntap_sys_set_hwaddr(struct device *dev, struct ether_addr *eth_addr) {
	return -1;
}

int
tuntap_sys_set_ipv4(struct device *dev, t_tun_in_addr *s4, uint32_t bits) {
	(void)dev;
	(void)s4;
	(void)bits;
	return -1;
}

int
tuntap_sys_set_ipv6(struct device *dev, t_tun_in6_addr *s6, uint32_t bits) {
	(void)dev;
	(void)s6;
	(void)bits;
	tuntap_log(TUNTAP_LOG_INFO, "IPv6 is not implemented on your system");
	return -1;
}

int
tuntap_sys_set_ifname(struct device *dev, const char *ifname, size_t len) {
	(void)dev;
	(void)ifname;
	(void)len;
	tuntap_log(TUNTAP_LOG_ERR,
	    "Your system does not support tuntap_set_ifname()");
	return -1;
}

int
tuntap_sys_set_descr(struct device *dev, const char *descr, size_t len) {
	(void)dev;
	(void)descr;
	(void)len;
	tuntap_log(TUNTAP_LOG_ERR,
	    "Your system does not support tuntap_set_descr()");
	return -1;
}

