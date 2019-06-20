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
	refreshInterval            = kingpin.Flag("refresh-interval", "Refresh interval").Default("0").Duration()
)

type Source struct {
	Name       string
	Identifier string
}

type Config struct {
	Sources           map[string][]Source
	StaticEnvironment map[string]string
}

func (c *Config) Refresh() (map[string]string, error) {
	env := map[string]string{}

	session, conf := common.OpenSession(flags)

	for secretType, sources := range c.Sources {
		switch secretType {
		case "SSM":
			ssmClient := ssm.New(session, conf)
			for _, source := range sources {
				if strings.HasSuffix(source.Identifier, "/*") {
					values, err := getParametersByPath(ssmClient, source.Identifier[:len(source.Identifier)-2], source.Name)
					if err != nil {
						return nil, err
					}
					for key, value := range values {
						env[key] = value
					}
				} else {
					value, err := ssmFetch(ssmClient, source.Identifier)
					if err != nil {
						return nil, err
					}
					env[source.Name] = value
				}
			}
		case "SECRETS_MANAGER":
			secretsManagerClient := secretsmanager.New(session, conf)
			for _, source := range sources {
				values, err := fetchSecret(secretsManagerClient, source.Identifier, source.Name)
				if err != nil {
					return nil, err
				}
				for key, value := range values {
					env[key] = value
				}
			}
		case "KMS":
			kmsClient := kms.New(session, conf)
			for _, source := range sources {
				value, err := kmsDecrypt(kmsClient, source.Identifier)
				if err != nil {
					return nil, err
				}
				env[source.Name] = value
			}
		}
	}

	for key, value := range c.StaticEnvironment {
		env[key] = value
	}

	return env, nil
}

func ParseConfig(env []string) (*Config, error) {
	config := &Config{
		Sources:           map[string][]Source{},
		StaticEnvironment: map[string]string{},
	}

	prefixes := map[string]string{
		"KMS":             *kmsPrefix,
		"SSM":             *ssmPrefix,
		"SECRETS_MANAGER": *secretsManagerPrefix,
	}
	for _, value := range env {
		parts := strings.SplitN(value, "=", 2)
		if len(parts) != 2 {
			continue
		}

		key := parts[0]
		value := parts[1]

		found := false
		for secretType, prefix := range prefixes {
			if strings.HasPrefix(key, prefix) {
				key := key[len(prefix):]

				name := key
				if strings.HasPrefix(key, "_") {
					name = ""
				}

				config.Sources[secretType] = append(config.Sources[secretType], Source{
					Name:       name,
					Identifier: value,
				})
				found = false
				break
			}
		}

		if !found {
			config.StaticEnvironment[key] = value
		}

	}

	return config, nil
}

func main() {
	kingpin.CommandLine.Name = "kms-env"
	kingpin.CommandLine.Help = "Decrypt environment variables encrypted with KMS, SSM or Secret Manager."
	kingpin.Parse()
	common.HandleInfoFlags(infoFlags)

	env := os.Environ()

	config, err := ParseConfig(env)
	common.FatalOnError(err)

	envMap, err := config.Refresh()
	common.FatalOnError(err)
	var pEnv []string
	for key, value := range envMap {
		pEnv = append(pEnv, fmt.Sprintf("%s=%s", key, value))
	}

	p := exec.Command((*command)[0], (*command)[1:]...)
	p.Env = pEnv
	p.Stdin = os.Stdin
	p.Stderr = os.Stderr
	p.Stdout = os.Stdout
	p.Run()

	os.Exit(p.ProcessState.ExitCode())
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
