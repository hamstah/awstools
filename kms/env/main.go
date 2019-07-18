package main

import (
	"bytes"
	"encoding/base64"
	"encoding/gob"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"reflect"
	"strings"
	"syscall"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/kms"
	"github.com/aws/aws-sdk-go/service/secretsmanager"
	"github.com/aws/aws-sdk-go/service/ssm"
	"github.com/hamstah/awstools/common"
	"golang.org/x/crypto/nacl/secretbox"
	kingpin "gopkg.in/alecthomas/kingpin.v2"
)

var (
	command                    = kingpin.Arg("command", "Command to run, prefix with -- to pass args").Required().Strings()
	kmsPrefix                  = kingpin.Flag("kms-prefix", "Prefix for the KMS environment variables").Default("KMS_").String()
	ssmPrefix                  = kingpin.Flag("ssm-prefix", "Prefix for the SSM environment variables").Default("SSM_").String()
	secretsManagerPrefix       = kingpin.Flag("secrets-manager-prefix", "Prefix for the secrets manager environment variables").Default("SECRETS_MANAGER_").String()
	secretsManagerVersionStage = kingpin.Flag("secrets-manager-version-stage", "The version stage of secrets from secrets manager").Default("AWSCURRENT").String()
	refreshInterval            = kingpin.Flag("refresh-interval", "Refresh interval").Default("0").Duration()
	refreshAction              = kingpin.Flag("refresh-action", "Action to take when values have changed").Default("RESTART").Enum("RESTART", "EXIT")
	refreshMaxRetries          = kingpin.Flag("refresh-max-retries", "Number of retries when failing to refresh the config").Default("5").Int()
)

type Source struct {
	Name       string
	Identifier string
}

type Config struct {
	Sources           map[string][]Source
	StaticEnvironment map[string]string
}

func (c *Config) IsRefreshable() bool {
	return len(c.Sources) > 0
}

func (c *Config) RefreshWithRetries(flags *common.SessionFlags) (map[string]string, error) {

	wait := 2

	for i := 0; i < *refreshMaxRetries; i++ {
		result, err := c.Refresh(flags)
		if err == nil {
			return result, nil
		}

		wait = wait * 2
		time.Sleep(time.Duration(wait) * time.Second)
	}
	return nil, errors.New("Failed to refresh config")
}

func (c *Config) Refresh(flags *common.SessionFlags) (map[string]string, error) {
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
					value, err := ssmGetParameter(ssmClient, source.Identifier)
					if err != nil {
						return nil, err
					}
					env[source.Name] = value
				}
			}
		case "SECRETS_MANAGER":
			secretsManagerClient := secretsmanager.New(session, conf)
			for _, source := range sources {
				values, err := secretsManagerGetSecretValue(secretsManagerClient, source.Identifier, source.Name)
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

func Monitor(flags *common.SessionFlags, config *Config, comm chan<- map[string]string) {
	previous, err := config.RefreshWithRetries(flags)
	if err != nil {
		comm <- nil
		return
	}
	comm <- previous

	if !config.IsRefreshable() || *refreshInterval == time.Duration(0) {
		return
	}

	for _ = range time.Tick(*refreshInterval) {
		new, err := config.RefreshWithRetries(flags)
		if err != nil {
			comm <- nil
			return
		}

		if !reflect.DeepEqual(new, previous) {
			comm <- new
		}
		previous = new
	}
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
	flags := common.HandleFlags()

	env := os.Environ()

	config, err := ParseConfig(env)
	common.FatalOnError(err)

	comm := make(chan map[string]string, 1)
	go Monitor(flags, config, comm)

	var p *exec.Cmd

	waitingPid := -1

	for envMap := range comm {
		if envMap == nil {
			// failed to refresh config, don't kill the process
			// stay alive even if config potentially out of date is better than a crash
			continue
		}

		if p != nil {
			waitingPid = p.Process.Pid
			p.Process.Signal(syscall.SIGTERM)
			p.Wait()

			if *refreshAction == "EXIT" {
				os.Exit(0)
			}
		}

		var pEnv []string
		for key, value := range envMap {
			pEnv = append(pEnv, fmt.Sprintf("%s=%s", key, value))
		}

		p = exec.Command((*command)[0], (*command)[1:]...)
		p.Env = pEnv
		p.Stdin = os.Stdin
		p.Stderr = os.Stderr
		p.Stdout = os.Stdout
		err := p.Start()
		common.FatalOnError(err)

		go func(p *exec.Cmd) {
			p.Wait()
			if waitingPid != p.Process.Pid {
				os.Exit(p.ProcessState.ExitCode())
			}
		}(p)
	}
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

func ssmGetParameter(ssmClient *ssm.SSM, name string) (string, error) {
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

func secretsManagerGetSecretValue(secretsManagerClient *secretsmanager.SecretsManager, secretName, prefix string) (map[string]string, error) {
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
