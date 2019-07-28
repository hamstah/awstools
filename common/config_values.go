package common

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/kms"
	"github.com/aws/aws-sdk-go/service/secretsmanager"
	"github.com/aws/aws-sdk-go/service/ssm"
	"github.com/aws/aws-sdk-go/service/sts"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

var SourceTypes = []string{"KMS", "SSM", "SECRETS_MANAGER", "FILE"}

type Source struct {
	Type       string
	Name       string
	Identifier string
	Collapse   bool
}

type ConfigValues struct {
	Sources       map[string][]Source
	Static        map[string]interface{}
	MaxRetries    int
	KeyPrefixes   map[string]string
	ValuePrefixes map[string]string
	Settings      map[string]string
}

func NewConfigValues() *ConfigValues {
	return &ConfigValues{
		Sources:    map[string][]Source{},
		Static:     map[string]interface{}{},
		MaxRetries: 5,
		KeyPrefixes: map[string]string{
			"KMS":             "KMS_",
			"SSM":             "SSM_",
			"SECRETS_MANAGER": "SECRETS_MANAGER_",
			"FILE":            "FILE_",
		},
		ValuePrefixes: map[string]string{
			"KMS":             "kms://",
			"SSM":             "ssm://",
			"SECRETS_MANAGER": "secrets-manager://",
			"FILE":            "file://",
		},
		Settings: map[string]string{
			"secrets_manager_version_stage": "AWSCURRENT",
		},
	}
}

func (c *ConfigValues) Clear() {
	c.Sources = map[string][]Source{}
	c.Static = map[string]interface{}{}
}

func (c *ConfigValues) SetFromJSON(filename string) error {
	value := map[string]interface{}{}
	err := LoadJSON(filename, &value)
	if err != nil {
		return err
	}
	return c.SetFromMap(value)
}

func (c *ConfigValues) SetFromEnvironment() error {
	value := map[string]interface{}{}
	for _, env := range os.Environ() {
		parts := strings.SplitN(env, "=", 2)
		if len(parts) != 2 {
			continue
		}
		value[parts[0]] = parts[1]
	}
	return c.SetFromMap(value)
}

func (c *ConfigValues) GenerateFromMap(src map[string]interface{}) (map[string]interface{}, error) {
	dst := map[string]interface{}{}

	for key, value := range src {
		switch value.(type) {
		case map[string]interface{}:
			val, err := c.GenerateFromMap(value.(map[string]interface{}))
			if err != nil {
				return nil, err
			}
			dst[key] = val
		case string:
			found := false

			for secretType, prefix := range c.KeyPrefixes {
				if strings.HasPrefix(key, prefix) {
					key := key[len(prefix):]
					collapse := strings.HasPrefix(key, "_")

					name := key
					if collapse && (secretType == "SECRETS_MANAGER" || secretType == "SSM") {
						name = name[1:]
					}

					dst[name] = Source{
						Type:       secretType,
						Name:       name,
						Identifier: value.(string),
						Collapse:   collapse,
					}
					found = true
					break
				}
			}

			if !found {
				collapse := strings.HasPrefix(key, "_")
				for secretType, prefix := range c.ValuePrefixes {
					if strings.HasPrefix(value.(string), prefix) {
						value := value.(string)[len(prefix):]

						name := key
						if collapse && (secretType == "SECRETS_MANAGER" || secretType == "SSM") {
							name = name[1:]
						}

						dst[name] = Source{
							Type:       secretType,
							Name:       name,
							Identifier: value,
							Collapse:   collapse,
						}
						found = true
						break
					}
				}
			}

			if !found {
				dst[key] = value
			}
		default:
			dst[key] = value
		}
	}

	return dst, nil
}

func (c *ConfigValues) SetFromMap(m map[string]interface{}) error {

	res, err := c.GenerateFromMap(m)
	if err != nil {
		return err
	}
	c.Static = res
	return nil
}

func (c *ConfigValues) IsRefreshable() bool {
	return len(c.Sources) > 0
}

func (c *ConfigValues) RefreshWithRetries(session *session.Session, conf *aws.Config, output interface{}) error {

	wait := 2

	for i := 0; i < c.MaxRetries; i++ {
		err := c.Refresh(session, conf, output)
		if err == nil {
			return nil
		}
		wait = wait * 2
		log.Error(errors.Wrap(err, fmt.Sprintf("Failed to refresh configuration, retrying in %ds", wait)))
		time.Sleep(time.Duration(wait) * time.Second)
	}
	return errors.New("Failed to refresh config")
}

type RefreshState struct {
	Session              *session.Session
	Config               *aws.Config
	STSClient            *sts.STS
	SecretsManagerClient *secretsmanager.SecretsManager
	KMSClient            *kms.KMS
	SSMClient            *ssm.SSM
	Settings             map[string]string
}

func (c *ConfigValues) Refresh(session *session.Session, conf *aws.Config, output interface{}) error {
	state := &RefreshState{
		Session:  session,
		Config:   conf,
		Settings: c.Settings,
	}
	env, err := RefreshMap(c.Static, state)
	if err != nil {
		return errors.Wrap(err, "failed to refresh config")
	}

	data, err := json.Marshal(env)
	if err != nil {
		return errors.Wrap(err, "failed to unmarshal the json config")
	}

	return json.Unmarshal(data, output)
}

func RefreshMap(src map[string]interface{}, state *RefreshState) (map[string]interface{}, error) {
	dst := map[string]interface{}{}

	for key, value := range src {
		switch value.(type) {
		case map[string]interface{}:
			res, err := RefreshMap(value.(map[string]interface{}), state)
			if err != nil {
				return nil, errors.Wrap(err, fmt.Sprintf("failed to refresh sub map with key %s", key))
			}
			dst[key] = res
		case Source:
			source := value.(Source)
			switch source.Type {
			case "FILE":
				bytes, err := ioutil.ReadFile(source.Identifier)
				if err != nil {
					return nil, errors.Wrap(err, fmt.Sprintf("failed to load file %s for key %s", source.Identifier, key))
				}
				dst[source.Name] = string(bytes)
			case "SSM":
				if state.SSMClient == nil {
					state.SSMClient = ssm.New(state.Session, state.Config)
				}
				if strings.HasSuffix(source.Identifier, "/*") {
					values, err := getParametersByPath(state.SSMClient, source.Identifier[:len(source.Identifier)-2])
					if err != nil {
						return nil, errors.Wrap(err, fmt.Sprintf("failed to fetch from SSM by path for key %s", key))
					}
					if source.Collapse {
						for subKey, subValue := range values {
							dst[subKey] = subValue
						}
					} else {
						dst[key] = values
					}
				} else {
					value, err := ssmGetParameter(state.SSMClient, source.Identifier)
					if err != nil {
						return nil, errors.Wrap(err, fmt.Sprintf("failed to fetch from SSM for key %s", key))
					}
					dst[source.Name] = value
				}
			case "SECRETS_MANAGER":
				if state.SecretsManagerClient == nil {
					state.SecretsManagerClient = secretsmanager.New(state.Session, state.Config)
				}
				values, err := secretsManagerGetSecretValue(state.SecretsManagerClient, source.Identifier, source.Name, state.Settings["secrets_manager_version_stage"])
				if err != nil {
					return nil, errors.Wrap(err, fmt.Sprintf("failed to fetch from Secrets Manager for key %s", key))
				}
				if source.Collapse {
					for subKey, subValue := range values {
						dst[subKey] = subValue
					}
				} else {
					dst[key] = values
				}
			case "KMS":
				if state.KMSClient == nil {
					state.KMSClient = kms.New(state.Session, state.Config)
				}
				value, err := DecryptWithKMS(state.KMSClient, source.Identifier)
				if err != nil {
					return nil, errors.Wrap(err, fmt.Sprintf("failed to decrypt KMS data for key %s", key))
				}
				dst[source.Name] = string(value)
			}
		default:
			dst[key] = value
		}
	}

	return dst, nil
}

func getParametersByPath(client *ssm.SSM, path string) (map[string]string, error) {
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
		key := parts[len(parts)-1]

		result[key] = *parameter.Value
	}

	return result, nil
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

func secretsManagerGetSecretValue(secretsManagerClient *secretsmanager.SecretsManager, secretName, prefix, versionStage string) (map[string]string, error) {
	result, err := secretsManagerClient.GetSecretValue(&secretsmanager.GetSecretValueInput{
		SecretId:     aws.String(secretName),
		VersionStage: aws.String(versionStage),
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
	return res, nil
}

func FlattenMap(input map[string]interface{}) (map[string]string, error) {
	result := map[string]string{}

	for key, value := range input {
		key = TransformKey(key)
		switch value.(type) {
		case int, string, float64, bool:
			result[key] = value.(string)
		case map[string]interface{}:
			sub, err := FlattenMap(value.(map[string]interface{}))
			if err != nil {
				return nil, err
			}
			for subKey, subValue := range sub {
				result[fmt.Sprintf("%s_%s", key, TransformKey(subKey))] = subValue
			}
		default:
			return nil, errors.New("Unsupported type")
		}
	}

	return result, nil
}

func TransformKey(key string) string {
	return strings.Replace(strings.ToUpper(key), "-", "_", -1)
}
