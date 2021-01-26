package main

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"os"
	"time"

	"github.com/aws/aws-sdk-go/service/lambda"
	"github.com/aws/aws-sdk-go/service/sts"
	"github.com/hamstah/awstools/common"
	"github.com/mitchellh/go-homedir"
	log "github.com/sirupsen/logrus"
	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/agent"
	kingpin "gopkg.in/alecthomas/kingpin.v2"
)

var (
	lambdaARN             = kingpin.Flag("lambda-arn", "ARN of the lambda function signing the SSH key.").Required().String()
	sshPrivateKeyFilename = kingpin.Flag("ssh-private-key-filename", "Path to the SSH key to add to the agent.").String()
	sshPublicKeyFilename  = kingpin.Flag("ssh-public-key-filename", "Path to the SSH key to sign.").String()
	environment           = kingpin.Flag("environment", "Name of the environment to sign the key for.").Default("").String()
	duration              = kingpin.Flag("duration", "Duration of validity of the signature.").Default("1m").Duration()
	dump                  = kingpin.Flag("dump", "Dump the event JSON instead of calling lambda.").Default("false").Bool()
	output                = kingpin.Flag("output", "Where to store the generated certificate.").Default("agent").Enum("agent", "stdout")
	sourceAddresses       = kingpin.Flag("source-address", "Set the IP restriction on the cert in CIDR format, can be repeated.").Strings()
	proxyConfig           = kingpin.Flag("proxy-config", "Configuration for the ssh ProxyCommand host:port.").String()
)

var (
	defaultSSHKeyLocations = []string{"~/.ssh/id_ecdsa", "~/.ssh/id_rsa"}
)

type SignSSHKeyResponse struct {
	Certificate string    `json:"certificate"`
	Duration    int       `json:"duration"`
	ValidBefore time.Time `json:"valid_before"`
}

type LambdaPayload struct {
	IdentityURL     string        `json:"identity_url"`
	Environment     string        `json:"environment"`
	SSHPublicKey    string        `json:"ssh_public_key"`
	Duration        time.Duration `json:"duration"`
	SourceAddresses []string      `json:"source_addresses"`
}

func HandleOptionalArgs() {
	if len(*sshPrivateKeyFilename) == 0 {
		found := false
		for _, defaultSSHKeyLocation := range defaultSSHKeyLocations {
			detected, err := homedir.Expand(defaultSSHKeyLocation)
			if err != nil {
				continue
			}

			if _, err := os.Stat(detected); err != nil {
				continue
			}

			*sshPrivateKeyFilename = detected
			found = true
			break
		}
		if !found {
			common.Fatalln("could not find the default private SSH key")
		}
		log.Info("Using default SSH key", *sshPrivateKeyFilename)
	}
	if len(*sshPublicKeyFilename) == 0 {
		*sshPublicKeyFilename = fmt.Sprintf("%s.pub", *sshPrivateKeyFilename)
	}
}

func ConnectIO(con net.Conn) {
	c := make(chan int64)

	copy := func(r io.ReadCloser, w io.WriteCloser) {
		defer func() {
			r.Close()
			w.Close()
		}()
		n, _ := io.Copy(w, r)
		c <- n
	}

	go copy(con, os.Stdout)
	go copy(os.Stdin, con)

	<-c
	<-c
}

func main() {
	kingpin.CommandLine.Name = "iam-request-ssh-key-signature"
	kingpin.CommandLine.Help = "Request a signature for a SSH key from lambda-sign-ssh-key."
	flags := common.HandleFlags()
	HandleOptionalArgs()

	sshPublicKeyBytes, err := ioutil.ReadFile(*sshPublicKeyFilename)
	common.FatalOnErrorW(err, "could not read the SSH public key")

	userSession, err := common.NewSession("")
	common.FatalOnErrorW(err, "could not open user IAM session")

	stsClient := sts.New(userSession)

	url, err := common.STSGetIdentityURL(stsClient)
	common.FatalOnErrorW(err, "could not generate the STS signed URL")

	session, conf := common.OpenSession(flags)

	lambdaPayload := LambdaPayload{
		IdentityURL:     url,
		Environment:     *environment,
		SSHPublicKey:    string(sshPublicKeyBytes),
		Duration:        *duration / time.Second,
		SourceAddresses: *sourceAddresses,
	}
	lambdaPayloadBytes, err := json.Marshal(lambdaPayload)
	common.FatalOnErrorW(err, "could not encode the lambda payload")

	if *dump {
		fmt.Println(string(lambdaPayloadBytes))
		os.Exit(0)
	}

	lambdaClient := lambda.New(session, conf)

	ret, err := lambdaClient.Invoke(&lambda.InvokeInput{
		FunctionName: lambdaARN,
		Payload:      lambdaPayloadBytes,
	})
	common.FatalOnErrorW(err, "could not invoke the lambda function")

	response := SignSSHKeyResponse{}
	err = json.Unmarshal(ret.Payload, &response)
	common.FatalOnErrorW(err, "could not parse the signature response")

	switch *output {
	case "agent":
		parsedCert, _, _, _, err := ssh.ParseAuthorizedKey([]byte(response.Certificate))
		common.FatalOnErrorW(err, "could not parse the certificate from the response")

		cert, ok := parsedCert.(*ssh.Certificate)
		if !ok {
			common.Fatalln("failed to parse response certificate")
		}

		privateKeyBytes, err := ioutil.ReadFile(*sshPrivateKeyFilename)
		common.FatalOnErrorW(err, "could not read the private SSH key")

		privateKey, err := ssh.ParseRawPrivateKey(privateKeyBytes)
		common.FatalOnErrorW(err, "could not parse the private SSH key")

		sock, err := net.Dial("unix", os.Getenv("SSH_AUTH_SOCK"))
		common.FatalOnErrorW(err, "could not connect to the SSH agent socket, make sure SSH_AUTH_SOCK is set")
		defer sock.Close()

		sshAgent := agent.NewClient(sock)

		pubcert := agent.AddedKey{
			Comment:      fmt.Sprintf("iam-request-ssh-key-signature env=\"%s\" expires=\"%s\"", *environment, response.ValidBefore),
			PrivateKey:   privateKey,
			Certificate:  cert,
			LifetimeSecs: uint32(response.Duration),
		}
		err = sshAgent.Add(pubcert)
		common.FatalOnErrorW(err, "could not add the certificate to the SSH agent")

	case "stdout":
		fmt.Println(string(ret.Payload))
	default:
	}

	if *proxyConfig != "" {
		conn, err := net.Dial("tcp", *proxyConfig)
		if err != nil {
			common.FatalOnErrorW(err, "could not connect to server")
		}
		ConnectIO(conn)
	}

}
