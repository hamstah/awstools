package main

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/kms"
	"github.com/aws/aws-sdk-go/service/ssm"
	"github.com/hamstah/awstools/common"
	"golang.org/x/crypto/nacl/secretbox"
	kingpin "gopkg.in/alecthomas/kingpin.v2"
)

var (
	flags     = common.KingpinSessionFlags()
	command   = kingpin.Arg("command", "Command to run, prefix with -- to pass args").Required().Strings()
	kmsPrefix = kingpin.Flag("kms-prefix", "Prefix for the KMS environment variables").Default("KMS_").String()
	ssmPrefix = kingpin.Flag("ssm-prefix", "Prefix for the SSM environment variables").Default("SSM_").String()
)

func main() {
	kingpin.CommandLine.Name = "kms_env"
	kingpin.CommandLine.Help = "Decrypt environment variables encrypted with KMS or SSM."
	kingpin.Parse()

	session, conf := common.OpenSession(flags)
	kmsClient := kms.New(session, conf)
	ssmClient := ssm.New(session, conf)

	env := os.Environ()
	var pEnv []string
	for _, value := range env {
		parts := strings.SplitN(value, "=", 2)
		if len(parts) != 2 {
			continue
		}

		newKey, newValue, err := handleEnvVar(kmsClient, ssmClient, parts[0], parts[1])
		common.FatalOnError(err)
		pEnv = append(pEnv, fmt.Sprintf("%s=%s", newKey, newValue))
	}

	p := exec.Command((*command)[0], (*command)[1:]...)
	p.Env = pEnv
	p.Stdin = os.Stdin
	p.Stderr = os.Stderr
	p.Stdout = os.Stdout
	p.Run()
}

func handleEnvVar(kmsClient *kms.KMS, ssmClient *ssm.SSM, key, value string) (string, string, error) {
	if strings.HasPrefix(key, *kmsPrefix) {
		newValue, err := kmsDecrypt(kmsClient, value)
		if err != nil {
			return "", "", err
		}
		return key[len(*kmsPrefix):], newValue, nil
	} else if strings.HasPrefix(key, *ssmPrefix) {
		newValue, err := ssmFetch(ssmClient, value)
		if err != nil {
			return "", "", err
		}
		return key[len(*ssmPrefix):], newValue, nil
	}

	return key, value, nil
}

const (
	keyLength   = 32
	nonceLength = 24
)

type payload struct {
	Key     []byte
	Nonce   *[nonceLength]byte
	Message []byte
}

func ssmFetch(ssmClient *ssm.SSM, name string) (string, error) {
	res, err := ssmClient.GetParameter(&ssm.GetParameterInput{
		Name:           aws.String(name),
		WithDecryption: aws.Bool(true),
	})
	if err != nil {
		return "", err
	}
	return *res.Parameter.Value, nil
}

func kmsDecrypt(kmsClient *kms.KMS, ciphertext string) (string, error) {
	// Decode ciphertext with gob
	var p payload
	gob.NewDecoder(bytes.NewReader([]byte(ciphertext))).Decode(&p)

	dataKeyOutput, err := kmsClient.Decrypt(&kms.DecryptInput{
		CiphertextBlob: p.Key,
	})
	if err != nil {
		return "", err
	}

	key := &[keyLength]byte{}
	copy(key[:], dataKeyOutput.Plaintext)

	// Decrypt message
	var plaintext []byte
	plaintext, ok := secretbox.Open(plaintext, p.Message, p.Nonce, key)
	if !ok {
		return "", fmt.Errorf("Failed to open secretbox")
	}
	return string(plaintext), nil
}
