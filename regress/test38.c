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

#include <stdio.h>
#if defined Windows
# include <windows.h>
#endif

#include "tuntap.h"

int debug, info, notice, warn, err;

void
test_cb(int level, const char *errmsg) {
	const char *prefix = NULL;

	switch (level) {
	case TUNTAP_LOG_DEBUG:
		prefix = "debug";
		debug = 1;
		break;
	case TUNTAP_LOG_INFO:
		prefix = "info";
		info = 1;
		break;
	case TUNTAP_LOG_NOTICE:
		prefix = "notice";
		notice = 1;
		break;
	case TUNTAP_LOG_WARN:
		prefix = "warn";
		warn = 1;
		break;
	case TUNTAP_LOG_ERR:
		prefix = "err";
		err = 1;
		break;
	default:
		/* NOTREACHED */
		break;
	}
	(void)fprintf(stderr, "%s: %s\n", prefix, errmsg);
}

int
main(void) {
	tuntap_log_set_cb(test_cb);

	tuntap_log(TUNTAP_LOG_DEBUG, "debug message");
	tuntap_log(TUNTAP_LOG_INFO, "info message");
	tuntap_log(TUNTAP_LOG_NOTICE, "notice message");
	tuntap_log(TUNTAP_LOG_WARN, "warn message");
	tuntap_log(TUNTAP_LOG_ERR, "err message");

	if (debug + info + notice + warn + err != 5)
		return -1;
	return 0;
}

