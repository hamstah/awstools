package common

import (
	"bytes"
	"crypto/rand"
	"encoding/base64"
	"encoding/gob"
	"fmt"

	"github.com/aws/aws-sdk-go/service/kms"
	"golang.org/x/crypto/nacl/secretbox"
)

const (
	keyLength   = 32
	nonceLength = 24
)

type payload struct {
	Key     []byte
	Nonce   *[nonceLength]byte
	Message []byte
}

func EncryptWithKMSAndSecretBox(kmsClient *kms.KMS, plaintext []byte, keyId string) (string, error) {
	keySpec := "AES_128"
	dataKeyInput := kms.GenerateDataKeyInput{KeyId: &keyId, KeySpec: &keySpec}

	dataKeyOutput, err := kmsClient.GenerateDataKey(&dataKeyInput)
	if err != nil {
		return "", err
	}
	// Initialize payload
	p := &payload{
		Key:   dataKeyOutput.CiphertextBlob,
		Nonce: &[nonceLength]byte{},
	}

	// Set nonce
	_, err = rand.Read(p.Nonce[:])
	if err != nil {
		return "", err
	}

	// Create key
	key := &[keyLength]byte{}
	copy(key[:], dataKeyOutput.Plaintext)

	// Encrypt message
	p.Message = secretbox.Seal(p.Message, plaintext, p.Nonce, key)

	buf := &bytes.Buffer{}
	err = gob.NewEncoder(buf).Encode(p)
	if err != nil {
		return "", err
	}
	output := base64.StdEncoding.EncodeToString(buf.Bytes())
	return output, nil
}

func DecryptWithKMS(kmsClient *kms.KMS, ciphertext string) ([]byte, error) {
	buf := make([]byte, base64.StdEncoding.DecodedLen(len(ciphertext)))
	l, err := base64.StdEncoding.Decode(buf, []byte(ciphertext))
	if err != nil {
		return nil, err
	}
	content := buf[:l]

	usingSecretBox := false
	if len(content) > 14 {
		if content[6] == 0x07 {
			if bytes.Equal(content[7:14], []byte("payload")) {
				usingSecretBox = true
			}
		}
	}

	var p payload

	if usingSecretBox {
		err = gob.NewDecoder(bytes.NewReader(content)).Decode(&p)
		if err != nil {
			return nil, err
		}
		content = p.Key
	}
	// fmt.Println(p)
	dataKeyOutput, err := kmsClient.Decrypt(&kms.DecryptInput{
		CiphertextBlob: content,
	})
	if err != nil {
		return nil, err
	}

	if usingSecretBox {
		key := &[keyLength]byte{}
		copy(key[:], dataKeyOutput.Plaintext)

		// Decrypt message
		var plaintext []byte
		plaintext, ok := secretbox.Open(plaintext, p.Message, p.Nonce, key)
		if !ok {
			return nil, fmt.Errorf("failed to open secretbox")
		}
		return plaintext, nil
	}

	return dataKeyOutput.Plaintext, nil
}
