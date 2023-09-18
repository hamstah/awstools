package main

import (
	"context"
	"encoding/json"
	"fmt"

	kingpin "github.com/alecthomas/kingpin/v2"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/hamstah/awstools/common"
)

var (
	configFilenameTemplate = kingpin.Flag("config-filename-template", "Filename of the configuration file.").Default("%s.json").String()
	eventFilename          = kingpin.Flag("event-filename", "Filename with the event payload. Will process the event and exit if present.").String()
	identityURLMaxAge      = kingpin.Flag("identity-url-max-age", "Maximum age of the identity URL signature.").Default("10s").Duration()
)

func main() {
	kingpin.CommandLine.Name = "lambda-sign-ssh-key"
	kingpin.CommandLine.Help = "Signs SSH keys."
	sessionFlags := common.HandleFlags()

	handler := Handler(sessionFlags, *configFilenameTemplate, *identityURLMaxAge)

	if *eventFilename == "" {
		lambda.Start(handler)
	} else {
		event := SignSSHKeyEvent{}
		err := common.LoadJSON(*eventFilename, &event)
		common.FatalOnError(err)

		response, err := handler(context.Background(), event)
		common.FatalOnError(err)

		bytes, err := json.MarshalIndent(response, "", "    ")
		common.FatalOnError(err)

		fmt.Println(string(bytes))
	}
}
