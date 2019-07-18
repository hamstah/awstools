package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"time"

	"github.com/aws/aws-sdk-go/service/lambda"
	"github.com/aws/aws-sdk-go/service/sts"
	"github.com/hamstah/awstools/common"
	kingpin "gopkg.in/alecthomas/kingpin.v2"
)

var (
	lambdaARN            = kingpin.Flag("lambda-arn", "ARN of the lambda function signing the SSH key.").Required().String()
	sshPublicKeyFilename = kingpin.Flag("ssh-public-key-filename", "Path to the SSH key to sign.").Required().String()
	environment          = kingpin.Flag("environment", "Name of the environment to sign the key for.").Default("").String()
	duration             = kingpin.Flag("duration", "Duration of validity of the signature.").Default("1m").Duration()
	dump                 = kingpin.Flag("dump", "Dump the event JSON instead of calling lambda").Default("false").Bool()
	sourceAddresses      = kingpin.Flag("source-address", "Set the IP restriction on the cert in CIDR format, can be repeated").Strings()
)

type LambdaPayload struct {
	IdentityURL     string        `json:"identity_url"`
	Environment     string        `json:"environment"`
	SSHPublicKey    string        `json:"ssh_public_key"`
	Duration        time.Duration `json:"duration"`
	SourceAddresses []string      `json:"source_addresses"`
}

func main() {
	kingpin.CommandLine.Name = "iam-request-ssh-key-signature"
	kingpin.CommandLine.Help = "Request a signature for a SSH key from lambda-sign-ssh-key."
	flags := common.HandleFlags()

	sshPublicKeyBytes, err := ioutil.ReadFile(*sshPublicKeyFilename)
	common.FatalOnError(err)

	userSession := common.NewSession("")
	stsClient := sts.New(userSession)

	url, err := common.STSGetIdentityURL(stsClient)
	common.FatalOnError(err)

	session, conf := common.OpenSession(flags)

	lambdaPayload := LambdaPayload{
		IdentityURL:     url,
		Environment:     *environment,
		SSHPublicKey:    string(sshPublicKeyBytes),
		Duration:        *duration / time.Second,
		SourceAddresses: *sourceAddresses,
	}
	lambdaPayloadBytes, err := json.Marshal(lambdaPayload)
	common.FatalOnError(err)

	if *dump {
		fmt.Println(string(lambdaPayloadBytes))
		os.Exit(0)
	}

	lambdaClient := lambda.New(session, conf)

	ret, err := lambdaClient.Invoke(&lambda.InvokeInput{
		FunctionName: lambdaARN,
		Payload:      lambdaPayloadBytes,
	})
	common.FatalOnError(err)
	fmt.Println(string(ret.Payload))
}
