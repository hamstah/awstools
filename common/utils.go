package common

import (
	"os/exec"
	"syscall"

	log "github.com/sirupsen/logrus"
)

func FatalOnError(err error) {
	if err != nil {
		log.Fatalln(err)
	}
}

func Fatalln(message string) {
	log.Fatalln(message)
}

func GetExitCode(cmd *exec.Cmd, err error) int {
	// adapted from https://stackoverflow.com/a/40770011
	if err != nil {
		// try to get the exit code
		if exitError, ok := err.(*exec.ExitError); ok {
			ws := exitError.Sys().(syscall.WaitStatus)
			return ws.ExitStatus()
		}
		return -1

	} else {
		// success, exitCode should be 0 if go is ok
		ws := cmd.ProcessState.Sys().(syscall.WaitStatus)
		return ws.ExitStatus()
	}
}
