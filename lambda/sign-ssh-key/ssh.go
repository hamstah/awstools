package main

import (
	"crypto/rand"
	"encoding/binary"
	"fmt"
	"strings"
	"time"

	"github.com/pkg/errors"
	"golang.org/x/crypto/ssh"
)

type Signer struct {
	CACert ssh.PublicKey
	CAKey  ssh.Signer

	AllowedClockDiff time.Duration
	MaxTTL           time.Duration
}

func (s *Signer) Init(key, passphrase, pubKey []byte, allowedClockDiff, maxTTL time.Duration) error {

	var err error
	if len(passphrase) != 0 {
		s.CAKey, err = ssh.ParsePrivateKeyWithPassphrase(key, passphrase)
	} else {
		s.CAKey, err = ssh.ParsePrivateKey(key)
	}
	if err != nil {
		return errors.Wrap(err, "error parsing CA private key")
	}

	s.CACert, _, _, _, err = ssh.ParseAuthorizedKey(pubKey)
	if err != nil {
		return errors.Wrap(err, "error parsing CA public key")
	}

	s.MaxTTL = maxTTL
	s.AllowedClockDiff = allowedClockDiff
	return nil
}

// ReadCA method read CA public cert from local file
func (s Signer) ReadCA() (string, error) {
	return string(ssh.MarshalAuthorizedKey(s.CACert)), nil
}

// Sign method is used to sign passed SSH Key.
func (s Signer) Sign(key []byte, keyId string, principals, sourceAddresses []string) (*ssh.Certificate, error) {

	buf := make([]byte, 8)
	_, err := rand.Read(buf)
	if err != nil {
		return nil, errors.Wrap(err, "failed to read random bytes")
	}
	serial := binary.LittleEndian.Uint64(buf)

	pubKey, _, _, _, err := ssh.ParseAuthorizedKey(key)
	if err != nil {
		return nil, fmt.Errorf("failed to parse user public key: %s", err)
	}

	criticalOptions := map[string]string{}
	if len(sourceAddresses) != 0 {
		criticalOptions["source-address"] = strings.Join(sourceAddresses, ",")
	}

	certificate := ssh.Certificate{
		Serial:          serial,
		Key:             pubKey,
		KeyId:           keyId,
		ValidPrincipals: principals,
		ValidAfter:      uint64(time.Now().UTC().Add(-s.AllowedClockDiff).Unix()),
		ValidBefore:     uint64(time.Now().UTC().Add(s.MaxTTL).Unix()),
		CertType:        ssh.UserCert,
		Permissions: ssh.Permissions{
			CriticalOptions: criticalOptions,
			Extensions: map[string]string{
				"permit-pty": "",
			},
		},
	}

	err = certificate.SignCert(rand.Reader, s.CAKey)
	if err != nil {
		return nil, fmt.Errorf("failed to sign user public key: %s", err)
	}

	return &certificate, nil
}
