package main

import (
	"context"
	"fmt"
	"time"

	"github.com/hamstah/awstools/common"
	"github.com/hashicorp/go-uuid"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"golang.org/x/crypto/ssh"
)

type SignSSHKeyResponse struct {
	Certificate string    `json:"certificate"`
	Duration    int       `json:"duration"`
	ValidBefore time.Time `json:"valid_before"`
}

type SignSSHKeyEvent struct {
	IdentityURL     string   `json:"identity_url"`
	Environment     string   `json:"environment"`
	SSHPublicKey    string   `json:"ssh_public_key"`
	Duration        *int     `json:"duration"`
	SourceAddresses []string `json:"source_addresses"`
}

func (event SignSSHKeyEvent) Validate() error {
	if event.IdentityURL == "" {
		return errors.New("'identity_url' not found in Event.")
	}

	if event.Environment == "" {
		return errors.New("'environment' not found in Event.")
	}

	if event.SSHPublicKey == "" {
		return errors.New("'ssh_public_key' not found in Event.")
	}

	if event.Duration != nil && *event.Duration < 1 {
		return errors.New("'duration' must be >= 1.")
	}

	// source address is validated later

	return nil
}

type HandlerFunc func(ctx context.Context, event SignSSHKeyEvent) (*SignSSHKeyResponse, error)

func Handler(sessionFlags *common.SessionFlags, configFilenameTemplate string, identityURLMaxAge time.Duration) HandlerFunc {
	return func(ctx context.Context, event SignSSHKeyEvent) (*SignSSHKeyResponse, error) {
		// check event
		err := event.Validate()
		if err != nil {
			return nil, errors.Wrap(err, "Invalid event")
		}

		// get config
		environment, err := LoadEnvironment(sessionFlags, configFilenameTemplate, event.Environment)
		if err != nil {
			return nil, errors.Wrap(err, "Failed to load the environment config")
		}

		// check source addresses
		sourceAddresses, err := ValidateIPRanges(event.SourceAddresses, environment.SourceAddresses)
		if err != nil {
			return nil, errors.Wrap(err, "Invalid source_addresses requested")
		}

		// check duration
		duration := environment.ValidityMaxDuration
		if event.Duration != nil {
			duration = event.Duration
		}

		if *duration > *environment.ValidityMaxDuration {
			return nil, errors.New("Requested duration exceeds the maximum duration for the environment.")
		}

		// check caller
		identity, err := common.STSFetchIdentityURL(event.IdentityURL, identityURLMaxAge)
		if err != nil {
			return nil, errors.Wrap(err, "Failed to verify caller identity")
		}

		userARN, err := common.ParseARN(*identity.Arn)
		if err != nil {
			return nil, errors.Wrap(err, "Failed to parse the identity ARN")
		}

		if userARN.ResourceType != "user" {
			return nil, errors.New("Caller identity should be an IAM user.")
		}

		signer := Signer{}
		err = signer.Init(
			[]byte(environment.CA.PrivateKey),
			[]byte(environment.CA.PrivateKeyPassphrase),
			[]byte(environment.CA.PublicKey),
			time.Duration(*environment.ValidityStartOffset)*time.Second,
			time.Duration(*duration)*time.Second,
		)
		if err != nil {
			return nil, errors.Wrap(err, "Failed to initialise signer")
		}

		keyUUID, err := uuid.GenerateUUID()
		if err != nil {
			return nil, errors.Wrap(err, "Failed to generate key UUID")
		}

		principals := []string{userARN.Resource}

		keyID := fmt.Sprintf("%s/%s", *identity.Arn, keyUUID)
		certificate, err := signer.Sign([]byte(event.SSHPublicKey), keyID, principals, sourceAddresses)
		if err != nil {
			return nil, errors.Wrap(err, "Failed to generate certificate")
		}

		marshaledCertificate := ssh.MarshalAuthorizedKey(certificate)
		if len(marshaledCertificate) == 0 {
			return nil, errors.New("failed to marshal signed certificate, empty result")
		}

		userIdentity := map[string]string{
			"type":        "IAMUser",
			"principalId": "",
			"arn":         *identity.Arn,
			"accountId":   userARN.AccountID,
			"accessKeyId": "",
			"userName":    userARN.Resource,
		}

		validBefore := time.Unix(int64(certificate.ValidBefore), 0).UTC()
		validAfter := time.Unix(int64(certificate.ValidAfter), 0).UTC()

		certificateLog := map[string]interface{}{
			"key_id":       keyID,
			"valid_before": validBefore,
			"valid_after":  validAfter,
			"principals":   principals,
		}

		log.WithFields(log.Fields{
			"userIdentity":   userIdentity,
			"ssh_public_key": event.SSHPublicKey,
			"environment":    event.Environment,
			"certificate":    certificateLog,
		}).Info("Generated SSH Key certificate")
		return &SignSSHKeyResponse{
			Certificate: string(marshaledCertificate),
			Duration:    *duration,
			ValidBefore: validBefore,
		}, nil
	}
}
