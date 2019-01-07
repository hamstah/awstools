package main

import (
	"github.com/aws/aws-sdk-go/service/kms"
)

func KMSListKeys(session *Session) *FetchResult {
	client := kms.New(session.Session, session.Config)

	result := &FetchResult{}
	result.Error = client.ListKeysPages(&kms.ListKeysInput{},
		func(page *kms.ListKeysOutput, lastPage bool) bool {
			for _, key := range page.Keys {

				resource, err := NewResource(*key.KeyArn)
				if err != nil {
					result.Error = err
					return false
				}

				describeResult, err := client.DescribeKey(&kms.DescribeKeyInput{KeyId: key.KeyId})
				if err != nil {
					result.Error = err
					return false
				}

				metadata := describeResult.KeyMetadata

				// Ignore default KMS keys
				if *metadata.KeyManager == kms.KeyManagerTypeAws {
					continue
				}

				// ignore deleted keys
				if *metadata.KeyState == kms.KeyStatePendingDeletion {
					continue
				}

				resource.Metadata = map[string]string{
					"Description": *metadata.Description,
					"KeyState":    *metadata.KeyState,
					"KeyUsage":    *metadata.KeyUsage,
				}
				result.Resources = append(result.Resources, *resource)
			}

			return true
		})

	return result
}
