package main

import (
	"encoding/base64"
	"fmt"
	"os"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ecr"
	kingpin "gopkg.in/alecthomas/kingpin.v2"
)

var (
	app    = kingpin.New("ecr-get-login", "Log into ECR docker registry")
	region = app.Flag("region", "AWS Region").Default("eu-west-1").String()
	output = app.Flag("output", "Return the credentials instead of docker command").Default("shell").Enum("raw", "shell")
)

func main() {

	kingpin.MustParse(app.Parse(os.Args[1:]))

	config := aws.Config{Region: aws.String(*region)}
	session := session.New(&config)
	svc := ecr.New(session)

	input := &ecr.GetAuthorizationTokenInput{}

	result, err := svc.GetAuthorizationToken(input)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Failed to get authorization token", err)
		os.Exit(2)
	}

	credentials := result.AuthorizationData[0]
	if *output == "raw" {
		fmt.Println(credentials)
	} else if *output == "shell" {

		data, err := base64.StdEncoding.DecodeString(*credentials.AuthorizationToken)
		if err != nil {
			fmt.Fprintln(os.Stderr, "Failed to decode the token", err)
			os.Exit(2)
		}

		parts := strings.SplitN(string(data), ":", 2)
		if len(parts) != 2 {
			fmt.Fprintln(os.Stderr, "Invalid token format")
			os.Exit(2)
		}

		fmt.Println(fmt.Sprintf("docker login -u %s -p %s %s", parts[0], parts[1], *credentials.ProxyEndpoint))
	}
}
