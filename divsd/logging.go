package divsd

import (
	logging "github.com/op/go-logging"
)

const LOG_MODULE = "divs"

var log = logging.MustGetLogger(LOG_MODULE)

// Example format string. Everything except the message has a custom color
// which is dependent on the log level. Many fields have a custom output
// formatting too, eg. the time returns the hour down to the milli second.
var format = "%{color}%{time:15:04:05.000000} ▶ %{level:.4s} %{id:03x}%{color:reset} %{message}"

func init() {
	logging.SetFormatter(logging.MustStringFormatter(format))

	// Setup one stderr and one syslog backend and combine them both into one
	// logging backend. By default stderr is used with the standard log flag.

	//logBackend := logging.NewLogBackend(os.Stderr, "", 0)
	//syslogBackend, err := logging.NewSyslogBackend("")
	//if err != nil {
	//	log.Fatal(err)
	//}
	//logging.SetBackend(logBackend, syslogBackend)

	// For "divs", set the log level to DEBUG and ERROR.
	//for _, level := range []logging.Level{logging.CRITICAL,
	//	logging.ERROR, logging.WARNING, logging.INFO} {
	//}
	logging.SetLevel(logging.INFO, LOG_MODULE)
}
