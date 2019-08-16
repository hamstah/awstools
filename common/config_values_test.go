package common

import (
	"encoding/json"
	"fmt"
	"os"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/kms"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	sess *session.Session
	conf *aws.Config
)

func TestMain(m *testing.M) {
	setup()
	os.Exit(m.Run())
}

func setup() {
	sess = session.Must(session.NewSessionWithOptions(session.Options{
		Profile:           "hamstah",
		SharedConfigState: session.SharedConfigEnable,
	}))
	conf = NewConfig("eu-west-1")
}

type Data struct {
	A string    `json:"a"`
	B float64   `json:"b"`
	C []string  `json:"c"`
	D []float64 `json:"d"`
	E struct {
		A string    `json:"a"`
		B float64   `json:"b"`
		C []string  `json:"c"`
		D []float64 `json:"d"`
	} `json:"e"`
}

func setFromMap(data string) (*ConfigValues, error) {
	c := NewConfigValues()
	m := map[string]interface{}{}
	err := json.Unmarshal([]byte(data), &m)
	if err != nil {
		return nil, err
	}

	err = c.SetFromMap(m)
	return c, err
}

func TestSetFromMap(t *testing.T) {

	data := `{
    "a": "123",
    "b": 456,
		"c": ["a", "b"],
		"d": [123, 456],
    "e": {
			"a": "123",
	    "b": 456,
			"c": ["a", "b"],
			"d": [123, 456]
		}
  }
  `
	c, err := setFromMap(data)
	require.NoError(t, err)

	res := Data{}
	err = c.Refresh(nil, nil, &res)

	assert.Equal(t, "123", res.A)
	assert.Equal(t, float64(456), res.B)
	assert.Equal(t, []string{"a", "b"}, res.C)
	assert.Equal(t, []float64{123, 456}, res.D)

	assert.Equal(t, "123", res.E.A)
	assert.Equal(t, float64(456), res.E.B)
	assert.Equal(t, []string{"a", "b"}, res.E.C)
	assert.Equal(t, []float64{123, 456}, res.E.D)
}

func TestSetFromMapWithValuePrefixes(t *testing.T) {
	data := `{
    "a": "abc",
    "b": "ssm:///secret-path",
    "c": {
      "d": "ssm:///secret-path",
      "e": "secrets-manager://secret-name",
      "f": "kms://value-to-decrypt",
      "g": "file://config_values.go",
      "h": "def",
      "i": 789
    }
  }
  `
	c, err := setFromMap(data)
	require.NoError(t, err)

	assert.Equal(t, Source{Type: "SSM", Name: "b", Identifier: "/secret-path"}, c.Static["b"])
	nested := c.Static["c"].(map[string]interface{})
	assert.Equal(t, Source{Type: "SSM", Name: "d", Identifier: "/secret-path"}, nested["d"])
	assert.Equal(t, Source{Type: "SECRETS_MANAGER", Name: "e", Identifier: "secret-name"}, nested["e"])
	assert.Equal(t, Source{Type: "KMS", Name: "f", Identifier: "value-to-decrypt"}, nested["f"])
	assert.Equal(t, Source{Type: "FILE", Name: "g", Identifier: "config_values.go"}, nested["g"])
	assert.Equal(t, "def", nested["h"])
	assert.Equal(t, float64(789), nested["i"])
}

type Struct1 struct {
	A string `json:"a"`
	B struct {
		C string `json:"c"`
	} `json:"b"`
}

func TestResolveSSMValuePrefixes(t *testing.T) {
	data := `{
    "a": "ssm:///hamstah/awstools/tests/test-1/key-1",
    "b": {
      "c": "ssm:///hamstah/awstools/tests/test-1/key-2"
    }
  }
  `
	c, err := setFromMap(data)
	require.NoError(t, err)

	d := Struct1{}
	err = c.Refresh(sess, conf, &d)
	require.NoError(t, err)

	assert.Equal(t, "value-1", d.A)
	assert.Equal(t, "value-2", d.B.C)
}

type SSM2 struct {
	A struct {
		Value1 string `json:"key-1"`
		Value2 string `json:"key-2"`
	} `json:"a"`
}

func TestResolveSSMValuePrefixesWildcard(t *testing.T) {
	data := `{
    "a": "ssm:///hamstah/awstools/tests/test-1/*"
  }
  `
	c, err := setFromMap(data)
	require.NoError(t, err)
	assert.Equal(t, Source{Type: "SSM", Name: "a", Identifier: "/hamstah/awstools/tests/test-1/*"}, c.Static["a"])

	d := SSM2{}
	err = c.Refresh(sess, conf, &d)
	require.NoError(t, err)

	assert.Equal(t, "value-1", d.A.Value1)
	assert.Equal(t, "value-2", d.A.Value2)
}

func TestResolveSSMKeyPrefixes(t *testing.T) {
	data := `{
    "SSM_a": "/hamstah/awstools/tests/test-1/key-1",
    "b": {
      "SSM_c": "/hamstah/awstools/tests/test-1/key-2"
    }
  }
  `
	c, err := setFromMap(data)
	require.NoError(t, err)

	d := Struct1{}
	err = c.Refresh(sess, conf, &d)
	require.NoError(t, err)

	assert.Equal(t, "value-1", d.A)
	assert.Equal(t, "value-2", d.B.C)
}

func TestResolveSSMKeyPrefixesWildcard(t *testing.T) {
	data := `{
    "SSM_a": "/hamstah/awstools/tests/test-1/*"
  }
  `
	c, err := setFromMap(data)
	require.NoError(t, err)
	assert.Equal(t, Source{Type: "SSM", Name: "a", Identifier: "/hamstah/awstools/tests/test-1/*"}, c.Static["a"])

	d := SSM2{}
	err = c.Refresh(sess, conf, &d)
	require.NoError(t, err)

	assert.Equal(t, "value-1", d.A.Value1)
	assert.Equal(t, "value-2", d.A.Value2)
}

type Struct2 struct {
	Value1 string `json:"key-1"`
	Value2 string `json:"key-2"`
	Value3 string `json:"key-3"`
	Value4 string `json:"key-4"`
}

func TestResolveSSMKeyPrefixesWildcardWithoutPrefix(t *testing.T) {
	data := `{
    "SSM__a": "/hamstah/awstools/tests/test-1/*"
  }
  `
	c, err := setFromMap(data)
	require.NoError(t, err)
	assert.Equal(t, Source{Type: "SSM", Name: "a", Identifier: "/hamstah/awstools/tests/test-1/*", Collapse: true}, c.Static["a"])

	d := Struct2{}
	err = c.Refresh(sess, conf, &d)
	require.NoError(t, err)

	assert.Equal(t, "value-1", d.Value1)
	assert.Equal(t, "value-2", d.Value2)
}

type SecretManager1 struct {
	A struct {
		Value1 string `json:"key-1"`
		Value2 string `json:"key-2"`
	} `json:"a"`
	B struct {
		C struct {
			Value1 string `json:"key-1"`
			Value2 string `json:"key-2"`
		} `json:"c"`
	} `json:"b"`
}

func TestResolveSecretsManagerValuePrefixes(t *testing.T) {
	data := `{
    "a": "secrets-manager://hamstah/awstools/tests/test-1",
    "b": {
      "c": "secrets-manager://hamstah/awstools/tests/test-1"
    }
  }
  `
	config, err := setFromMap(data)
	require.NoError(t, err)

	d := SecretManager1{}
	err = config.Refresh(sess, conf, &d)
	require.NoError(t, err)

	assert.Equal(t, "value-1", d.A.Value1)
	assert.Equal(t, "value-2", d.A.Value2)

	assert.Equal(t, "value-1", d.B.C.Value1)
	assert.Equal(t, "value-2", d.B.C.Value2)
}

func TestResolveSecretsManagerValuePrefixWithoutPrefix(t *testing.T) {
	data := `{
		"_a": "secrets-manager://hamstah/awstools/tests/test-1"
  	}
	  `

	c, err := setFromMap(data)
	require.NoError(t, err)
	assert.Equal(t, Source{Type: "SECRETS_MANAGER", Name: "a", Identifier: "hamstah/awstools/tests/test-1", Collapse: true}, c.Static["a"])

	d := Struct2{}
	err = c.Refresh(sess, conf, &d)
	require.NoError(t, err)

	assert.Equal(t, "value-1", d.Value1)
	assert.Equal(t, "value-2", d.Value2)
}

func TestSecretsManagerWithoutPrefix(t *testing.T) {

	data := `{
		"SECRETS_MANAGER__a": "hamstah/awstools/tests/test-1",
		"_B": "secrets-manager://hamstah/awstools/tests/test-2"
	}
	`
	config, err := setFromMap(data)
	require.NoError(t, err)

	fmt.Println(config.Static)
	assert.Equal(t, Source{Type: "SECRETS_MANAGER", Name: "a", Identifier: "hamstah/awstools/tests/test-1", Collapse: true}, config.Static["a"])
	assert.Equal(t, Source{Type: "SECRETS_MANAGER", Name: "B", Identifier: "hamstah/awstools/tests/test-2", Collapse: true}, config.Static["B"])

	d := Struct2{}
	err = config.Refresh(sess, conf, &d)
	require.NoError(t, err)

	assert.Equal(t, "value-1", d.Value1)
}

func TestResolveSecretsManagerKeyPrefixes(t *testing.T) {
	data := `{
    "SECRETS_MANAGER_a": "hamstah/awstools/tests/test-1",
    "b": {
      "SECRETS_MANAGER_c": "hamstah/awstools/tests/test-1"
    }
  }
  `
	config, err := setFromMap(data)
	require.NoError(t, err)

	d := SecretManager1{}
	err = config.Refresh(sess, conf, &d)
	require.NoError(t, err)

	assert.Equal(t, "value-1", d.A.Value1)
	assert.Equal(t, "value-2", d.A.Value2)

	assert.Equal(t, "value-1", d.B.C.Value1)
	assert.Equal(t, "value-2", d.B.C.Value2)
}

func TestResolveKMSKeyPrefixes(t *testing.T) {
	kmsValue := "AQICAHgn4TiP+IVP5oaT3N1aUybKUHg7vUly/WosR5LPrHVjnQEMLudv32oTAQWoX6HniNdEAAAAYzBhBgkqhkiG9w0BBwagVDBSAgEAME0GCSqGSIb3DQEHATAeBglghkgBZQMEAS4wEQQM/V5L24Ql0EtrGRgKAgEQgCArH5vr2jJJJErnnv2o/xuf/eKBg2fdVFcX92hbcKi1Ng=="
	data := fmt.Sprintf(`{
    "KMS_a": "%s",
    "b": {
      "KMS_c": "%s"
    }
  }
  `, kmsValue, kmsValue)
	config, err := setFromMap(data)
	require.NoError(t, err)

	d := Struct1{}
	err = config.Refresh(sess, conf, &d)
	require.NoError(t, err)

	assert.Equal(t, "value", d.A)
	assert.Equal(t, "value", d.B.C)
}

func TestResolveKMSSecretBoxKeyPrefixes(t *testing.T) {
	kmsValue := "NP+BAwEBB3BheWxvYWQB/4IAAQMBA0tleQEKAAEFTm9uY2UB/4QAAQdNZXNzYWdlAQoAAAAZ/4MBAQEJWzI0XXVpbnQ4Af+EAAEGATAAAP/r/4IB/6gBAgMAeCfhOI/4hU/mhpPc3VpTJspQeDu9SXL9aixHks+sdWOdAScFLzbtavfzeThGQcMxSC8AAABuMGwGCSqGSIb3DQEHBqBfMF0CAQAwWAYJKoZIhvcNAQcBMB4GCWCGSAFlAwQBLjARBAyhDR1pFVAkoFoCVWYCARCAKw5BfzV6W31ZYNzIYgcT0+LI5pEekbXSf09o7AuIrSs5yu4LI/waNDH2P94BGG3/6XxeAwF2Wwn/0/+7/6j/5/+0ACr/5/+e/7D/7nf/kEr/wwEVKcbN1sW5fOl3/A2JPW+dDLXC3+dPAA=="
	data := fmt.Sprintf(`{
    "KMS_a": "%s",
    "b": {
      "KMS_c": "%s"
    }
  }
  `, kmsValue, kmsValue)
	config, err := setFromMap(data)
	require.NoError(t, err)

	d := Struct1{}
	err = config.Refresh(sess, conf, &d)
	require.NoError(t, err)

	assert.Equal(t, "value", d.A)
	assert.Equal(t, "value", d.B.C)
}

func TestEncryptDecryptKMSWithSecretBox(t *testing.T) {
	kmsClient := kms.New(sess, conf)
	plaintext := []byte("value")
	keyId := "41694ce4-3596-455c-89ae-b36f9e20566a"

	encrypted, err := EncryptWithKMSAndSecretBox(kmsClient, plaintext, keyId)
	require.NoError(t, err)

	decrypted, err := DecryptWithKMS(kmsClient, encrypted)
	require.NoError(t, err)
	assert.Equal(t, plaintext, decrypted)
}
