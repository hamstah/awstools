package common

import kingpin "github.com/alecthomas/kingpin/v2"

func HandleFlags() *SessionFlags {
	sessionFlags := KingpinSessionFlags()
	infoFlags := KingpinInfoFlags()
	logFlags := KingpinLogFlags()

	kingpin.Parse()
	HandleInfoFlags(infoFlags)
	HandleLogFlags(logFlags)
	return sessionFlags
}
