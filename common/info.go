package common

import (
	"fmt"
	"os"

	kingpin "github.com/alecthomas/kingpin/v2"
)

var (
	Version    string = "???"
	CommitHash string = "???"
)

type InfoFlags struct {
	Version *bool
}

func VersionString() string {
	return fmt.Sprintf("%s (%s)", Version, CommitHash)
}

func HandleInfoFlags(flags *InfoFlags) {
	if *flags.Version {
		fmt.Println(VersionString())
		os.Exit(0)
	}
}

func KingpinInfoFlags() *InfoFlags {

	return &InfoFlags{
		Version: kingpin.Flag("version", "Display the version").Short('v').Bool(),
	}
}
