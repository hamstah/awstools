package main

import (
	"bytes"
	"encoding/base64"
	"encoding/gob"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/kms"
	"github.com/aws/aws-sdk-go/service/secretsmanager"
	"github.com/aws/aws-sdk-go/service/ssm"
	"github.com/hamstah/awstools/common"
	"golang.org/x/crypto/nacl/secretbox"
	kingpin "gopkg.in/alecthomas/kingpin.v2"
)

var (
	flags                      = common.KingpinSessionFlags()
	infoFlags                  = common.KingpinInfoFlags()
	command                    = kingpin.Arg("command", "Command to run, prefix with -- to pass args").Required().Strings()
	kmsPrefix                  = kingpin.Flag("kms-prefix", "Prefix for the KMS environment variables").Default("KMS_").String()
	ssmPrefix                  = kingpin.Flag("ssm-prefix", "Prefix for the SSM environment variables").Default("SSM_").String()
	secretsManagerPrefix       = kingpin.Flag("secrets-manager-prefix", "Prefix for the secrets manager environment variables").Default("SECRETS_MANAGER_").String()
	secretsManagerVersionStage = kingpin.Flag("secrets-manager-version-stage", "The version stage of secrets from secrets manager").Default("AWSCURRENT").String()
)

func main() {
	kingpin.CommandLine.Name = "kms-env"
	kingpin.CommandLine.Help = "Decrypt environment variables encrypted with KMS, SSM or Secret Manager."
	kingpin.Parse()
	common.HandleInfoFlags(infoFlags)

	session, conf := common.OpenSession(flags)
	kmsClient := kms.New(session, conf)
	ssmClient := ssm.New(session, conf)
	secretsManagerClient := secretsmanager.New(session, conf)

	env := os.Environ()
	var pEnv []string
	for _, value := range env {
		parts := strings.SplitN(value, "=", 2)
		if len(parts) != 2 {
			continue
		}

		result, err := handleEnvVar(kmsClient, ssmClient, secretsManagerClient, parts[0], parts[1])
		common.FatalOnError(err)
		for newKey, newValue := range result {
			pEnv = append(pEnv, fmt.Sprintf("%s=%s", newKey, newValue))
		}
	}

	p := exec.Command((*command)[0], (*command)[1:]...)
	p.Env = pEnv
	p.Stdin = os.Stdin
	p.Stderr = os.Stderr
	p.Stdout = os.Stdout
	p.Run()
}

func handleEnvVar(kmsClient *kms.KMS, ssmClient *ssm.SSM, secretsManagerClient *secretsmanager.SecretsManager, key, value string) (map[string]string, error) {
	if strings.HasPrefix(key, *kmsPrefix) {
		newValue, err := kmsDecrypt(kmsClient, value)
		if err != nil {
			return nil, err
		}
		return map[string]string{key[len(*kmsPrefix):]: newValue}, nil
	} else if strings.HasPrefix(key, *ssmPrefix) {
		if strings.HasSuffix(value, "/*") {
			prefix := key[len(*ssmPrefix):]
			// remove prefix if starts with _
			if strings.HasPrefix(prefix, "_") {
				prefix = ""
			}
			return getParametersByPath(ssmClient, value[:len(value)-2], prefix)
		} else {
			newValue, err := ssmFetch(ssmClient, value)
			if err != nil {
				return nil, err
			}
			return map[string]string{key[len(*ssmPrefix):]: newValue}, nil
		}
	} else if strings.HasPrefix(key, *secretsManagerPrefix) {
		prefix := key[len(*secretsManagerPrefix):]
		if strings.HasPrefix(prefix, "_") {
			prefix = ""
		}
		return fetchSecret(secretsManagerClient, value, prefix)
	}

	return map[string]string{key: value}, nil
}

func getParametersByPath(client *ssm.SSM, path string, prefix string) (map[string]string, error) {
	res, err := client.GetParametersByPath(&ssm.GetParametersByPathInput{
		Path:           aws.String(path),
		WithDecryption: aws.Bool(true),
	})
	if err != nil {
		return nil, err
	}
	result := map[string]string{}

	for _, parameter := range res.Parameters {
		parts := strings.Split(*parameter.Name, "/")
		key := strings.Replace(parts[len(parts)-1], "-", "_", -1)
		key = strings.ToUpper(key)
		if prefix != "" {
			key = fmt.Sprintf("%s_%s", prefix, key)
		}
		result[key] = *parameter.Value
	}

	return result, nil
}

func ConvertMap(source map[string]string, prefix string) map[string]string {
	res := make(map[string]string, len(source))
	for key, value := range source {
		var newKey string
		if prefix == "" {
			newKey = strings.ToUpper(key)
		} else {
			newKey = fmt.Sprintf("%s_%s", prefix, strings.ToUpper(key))
		}
		res[newKey] = value
	}
	return res
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

func fetchSecret(secretsManagerClient *secretsmanager.SecretsManager, secretName, prefix string) (map[string]string, error) {
	result, err := secretsManagerClient.GetSecretValue(&secretsmanager.GetSecretValueInput{
		SecretId:     aws.String(secretName),
		VersionStage: secretsManagerVersionStage,
	})
	if err != nil {
		return nil, err
	}

	var content []byte
	if result.SecretString != nil {
		content = []byte(*result.SecretString)
	} else {
		decodedBinarySecretBytes := make([]byte, base64.StdEncoding.DecodedLen(len(result.SecretBinary)))
		len, err := base64.StdEncoding.Decode(decodedBinarySecretBytes, result.SecretBinary)
		if err != nil {
			return nil, err
		}
		content = decodedBinarySecretBytes[:len]
	}

	res := make(map[string]string)
	err = json.Unmarshal(content, &res)
	if err != nil {
		return nil, err
	}

	return ConvertMap(res, prefix), nil
}
