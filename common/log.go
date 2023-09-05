package common

import (
	kingpin "github.com/alecthomas/kingpin/v2"
	log "github.com/sirupsen/logrus"
)

type LogFlags struct {
	LogLevel  *string
	LogFormat *string
}

func HandleLogFlags(flags *LogFlags) {
	switch *flags.LogFormat {
	case "json":
		log.SetFormatter(&log.JSONFormatter{})
	case "text":
		log.SetFormatter(&log.TextFormatter{})
	default:
		Fatalln("Invalid --log-format value")
	}

	switch *flags.LogLevel {
	case "trace":
		log.SetLevel(log.TraceLevel)
	case "debug":
		log.SetLevel(log.DebugLevel)
	case "info":
		log.SetLevel(log.InfoLevel)
	case "warn":
		log.SetLevel(log.WarnLevel)
	case "error":
		log.SetLevel(log.ErrorLevel)
	case "fatal":
		log.SetLevel(log.FatalLevel)
	case "panic":
		log.SetLevel(log.PanicLevel)
	default:
		Fatalln("Invalid --log-level")
	}
}

func KingpinLogFlags() *LogFlags {

	return &LogFlags{
		LogLevel:  kingpin.Flag("log-level", "Log level").Default("warn").Enum("trace", "debug", "info", "warn", "error", "fatal", "panic"),
		LogFormat: kingpin.Flag("log-format", "Log format").Default("text").Enum("text", "json"),
	}
}
