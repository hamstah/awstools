package main

import (
	"strings"

	"github.com/aws/aws-sdk-go/service/kms"
	"github.com/fatih/structs"
)

var (
	KMSService = Service{
		Name: "kms",
		Reports: map[string]Report{
			"keys":    KMSListKeys,
			"aliases": KMSListAliases,
		},
	}
)

func KMSListKeys(session *Session) *ReportResult {
	client := kms.New(session.Session, session.Config)

	result := &ReportResult{}
	result.Error = client.ListKeysPages(&kms.ListKeysInput{},
		func(page *kms.ListKeysOutput, lastPage bool) bool {
			for _, key := range page.Keys {

				resource, err := NewResource(*key.KeyArn, key)
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

				resource.Metadata = structs.Map(metadata)
				result.Resources = append(result.Resources, *resource)
			}

			return true
		})

	return result
}

func KMSListAliases(session *Session) *ReportResult {
	client := kms.New(session.Session, session.Config)

	result := &ReportResult{}
	result.Error = client.ListAliasesPages(&kms.ListAliasesInput{},
		func(page *kms.ListAliasesOutput, lastPage bool) bool {
			for _, alias := range page.Aliases {

				if strings.HasPrefix(*alias.AliasName, "alias/aws/") {
					continue
				}

				resource, err := NewResource(*alias.AliasArn, alias)
				if err != nil {
					result.Error = err
					return false
				}
				result.Resources = append(result.Resources, *resource)
			}

			return true
		})

	return result
}
