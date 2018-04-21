package main

import (
	"encoding/base64"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ecr"
	"github.com/hamstah/awstools/common"
	kingpin "gopkg.in/alecthomas/kingpin.v2"
)

var (
	flags  = common.KingpinSessionFlags()
	output = kingpin.Flag("output", "Return the credentials instead of docker command").Default("shell").Enum("raw", "shell")
)

func main() {
	kingpin.CommandLine.Name = "ecr-get-login"
	kingpin.CommandLine.Help = "Returns an authorization token from ECR."
	kingpin.Parse()

	session := session.Must(session.NewSession())
	conf := common.AssumeRoleConfig(flags, session)

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

		fmt.Println(fmt.Sprintf("docker login -u %s -p %s %s", parts[0], parts[1], *credentials.ProxyEndpoint))
	}
}
