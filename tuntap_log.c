/**
 * Copyright (c) 2012, PICHOT Fabien Paul Leonard <pichot.fabienATgmail.com>
 * Copyright (c) 2012, Tristan Le Guern <leguern AT medu DOT se>
 *
 * Permission to use, copy, modify, and/or distribute this software for any
 * purpose with or without fee is hereby granted, provided that the above
 * copyright notice and this permission notice appear in all copies.
 *
 * THE SOFTWARE IS PROVIDED "AS IS" AND THE AUTHOR DISCLAIMS ALL WARRANTIES WITH
 * REGARD TO THIS SOFTWARE INCLUDING ALL IMPLIED WARRANTIES OF MERCHANTABILITY
 * AND FITNESS. IN NO EVENT SHALL THE AUTHOR BE LIABLE FOR ANY SPECIAL, DIRECT,
 * INDIRECT, OR CONSEQUENTIAL DAMAGES OR ANY DAMAGES WHATSOEVER RESULTING FROM
 * LOSS OF USE, DATA OR PROFITS, WHETHER IN AN ACTION OF CONTRACT, NEGLIGENCE OR
 * OTHER TORTIOUS ACTION, ARISING OUT OF OR IN CONNECTION WITH THE USE OR
 * PERFORMANCE OF THIS SOFTWARE.
**/

#if defined Windows
# include <windows.h>
#endif
#include <stdio.h>
#include <string.h>
#include <ctype.h>
#include <stdint.h>

#include "tuntap.h"

void
tuntap_log_set_cb(t_tuntap_log cb) {
    if (cb == NULL)
        tuntap_log = tuntap_log_default;
	tuntap_log = cb;
}

void
tuntap_log_default(int level, const char *errmsg) {
	char *name;

	switch(level) {
	case TUNTAP_LOG_DEBUG:
		name = "Debug";
		break;
	case TUNTAP_LOG_INFO:
		name = "Info";
		break;
	case TUNTAP_LOG_NOTICE:
		name = "Notice";
		break;
	case TUNTAP_LOG_WARN:
		name = "Warning";
		break;
	case TUNTAP_LOG_ERR:
		name = "Error";
		break;
	case TUNTAP_LOG_NONE:
	default:
		name = NULL;
		break;
	}
	if (name == NULL) {
		(void)fprintf(stderr, "%s\n", errmsg);
	} else {
		(void)fprintf(stderr, "%s: %s\n", name, errmsg);
	}
}

void
tuntap_log_hexdump(void *data, size_t size) {
	unsigned char *p = (unsigned char *)data;
	unsigned int c;
	size_t n;
	char bytestr[4] = {0};
	char addrstr[10] = {0};
	char hexstr[16 * 3 + 5] = {0};
	char charstr[16 * 1 + 5] = {0};
	char buf[1024];

	for (n = 1; n <= size; n++) {
		if (n % 16 == 1) {
			/* store address for this line */
			snprintf(addrstr, sizeof(addrstr), "%.4lx",
			    ((uintptr_t)p - (uintptr_t)data) );
		}

		c = *p;
		if (isalnum(c) == 0) {
			c = '.';
		}

		/* store hex str (for left side) */
		snprintf(bytestr, sizeof(bytestr), "%02X ", *p);
		strncat(hexstr, bytestr, sizeof(hexstr)-strlen(hexstr)-1);

		/* store char str (for right side) */
		snprintf(bytestr, sizeof(bytestr), "%c", c);
		strncat(charstr, bytestr, sizeof(charstr)-strlen(charstr)-1);

		if (n % 16 == 0) { 
			/* line completed */
			(void)memset(buf, 0, sizeof buf);
			(void)snprintf(buf, sizeof buf,
			    "[%4.4s]   %-50.50s  %s", addrstr, hexstr, charstr);
			tuntap_log(TUNTAP_LOG_NONE, buf);
			hexstr[0] = 0;
			charstr[0] = 0;
		} else if (n % 8 == 0) {
			/* half line: add whitespaces */
			strncat(hexstr, "  ", sizeof(hexstr)-strlen(hexstr)-1);
			strncat(charstr, " ", sizeof(charstr)-strlen(charstr)-1);
		}
		p++; /* next byte */
	}

	/* print the rest of the buffer if not empty */
	if (strlen(hexstr) > 0) {
		(void)memset(buf, 0, sizeof buf);
		(void)snprintf(buf, sizeof buf, "[%4.4s]   %-50.50s  %s",
				addrstr, hexstr, charstr);
		tuntap_log(TUNTAP_LOG_NONE, buf);
	}
}

void
tuntap_log_chksum(void *addr, int count) {
	int sum;
	short *sptr;
	char buf[32];
	
	sum = 0;
	sptr = (short *)addr;
	while (count > 1)
	{
		sum = sum + *sptr;
		count = count - 2;
		sptr++;
	}
	
	addr = (char *)sptr;
	if (count > 0)
		sum = sum + *((char *) addr);
	sum = ~sum;

	(void)memset(buf, 0, sizeof buf);
	(void)snprintf(buf, sizeof buf, "Checksum of this block: %0#4x", sum);
	tuntap_log(TUNTAP_LOG_NONE, buf);
}

