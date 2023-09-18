package main

import (
	"encoding/base64"
	"fmt"
	"strings"

	kingpin "github.com/alecthomas/kingpin/v2"
	"github.com/aws/aws-sdk-go/service/ecr"
	"github.com/hamstah/awstools/common"
)

var (
	output = kingpin.Flag("output", "Return the credentials instead of docker command").Default("shell").Enum("raw", "shell")
)

func main() {
	kingpin.CommandLine.Name = "ecr-get-login"
	kingpin.CommandLine.Help = "Returns an authorization token from ECR."
	flags := common.HandleFlags()

	session, conf := common.OpenSession(flags)

	ecrClient := ecr.New(session, conf)
	result, err := ecrClient.GetAuthorizationToken(&ecr.GetAuthorizationTokenInput{})
	common.FatalOnError(err)

	credentials := result.AuthorizationData[0]
	if *output == "raw" {
		fmt.Println(credentials)
	} else if *output == "shell" {
		data, err := base64.StdEncoding.DecodeString(*credentials.AuthorizationToken)
		common.FatalOnError(err)

		parts := strings.SplitN(string(data), ":", 2)
		if len(parts) != 2 {
			common.Fatalln("Invalid token format")
		}

		fmt.Printf("docker login -u %s -p %s %s\n", parts[0], parts[1], *credentials.ProxyEndpoint)
	}
}
