package common

import kingpin "gopkg.in/alecthomas/kingpin.v2"

func HandleFlags() *SessionFlags {
	sessionFlags := KingpinSessionFlags()
	infoFlags := KingpinInfoFlags()
	logFlags := KingpinLogFlags()

	kingpin.Parse()
	HandleInfoFlags(infoFlags)
	HandleLogFlags(logFlags)
	return sessionFlags
}
