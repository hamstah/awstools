package main

import (
	"encoding/base64"
	"fmt"

	"github.com/hamstah/awstools/common"
	"github.com/pkg/errors"
)

var (
	DefaultValidityMaxDuration = 60
	DefaultValidityStartOffset = 60
)

type CA struct {
	PrivateKeyPassphrase string `json:"private_key_passphrase"`
	PrivateKey           string `json:"private_key"`
	PublicKey            string `json:"public_key"`
}

type Environment struct {
	CA                  CA                `json:"ca"`
	ValidityMaxDuration *int              `json:"validity_max_duration"`
	ValidityStartOffset *int              `json:"validity_start_offset"`
	Extensions          map[string]string `json:"extensions"`
	SourceAddresses     []string          `json:"source_addreses"`
}

func (e *Environment) Validate() error {
	if e.ValidityMaxDuration == nil {
		e.ValidityMaxDuration = &DefaultValidityMaxDuration
	}

	if e.ValidityStartOffset == nil {
		e.ValidityStartOffset = &DefaultValidityStartOffset
	}

	if e.CA.PublicKey == "" {
		return errors.New("'ca_public_key' is missing in environment config.")
	}

	if e.CA.PrivateKey == "" {
		return errors.New("'ca_private_key' is missing in environment config.")
	} else {
		data, err := base64.StdEncoding.DecodeString(e.CA.PrivateKey)
		if err == nil {
			e.CA.PrivateKey = string(data)
		}
	}

	if e.Extensions == nil {
		e.Extensions = map[string]string{}
	}

	return nil
}

func LoadEnvironment(sessionFlags *common.SessionFlags, configFilenameTemplate, environmentName string) (*Environment, error) {

	c := common.NewConfigValues()
	err := c.SetFromJSON(fmt.Sprintf(configFilenameTemplate, environmentName))
	if err != nil {
		return nil, err
	}

	session, config := common.OpenSession(sessionFlags)

	environment := &Environment{}
	err = c.Refresh(session, config, environment)
	if err != nil {
		return nil, err
	}
	environment.Validate()

	return environment, nil
}
