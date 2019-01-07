package main

import (
	"encoding/json"
	"io/ioutil"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sts"
	"github.com/hamstah/awstools/common"
)

type Account struct {
	Regions     []string `json:"regions"`
	RoleARN     string   `json:"role_arn"`
	ExternalID  string   `json:"external_id"`
	SessionName string   `json:"session_name"`
}

type Accounts struct {
	Accounts []*Account `json:"accounts"`
	Sessions []*Session
}

type Session struct {
	Session   *session.Session
	Config    *aws.Config
	AccountID string
}

func NewAccounts(filename string) (*Accounts, error) {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	result := &Accounts{}
	err = json.Unmarshal(data, result)
	if err != nil {
		return nil, err
	}

	for _, account := range result.Accounts {
		for _, region := range account.Regions {
			sess, conf := common.OpenSession(&common.SessionFlags{
				RoleArn:         &account.RoleARN,
				RoleExternalID:  &account.ExternalID,
				Region:          &region,
				RoleSessionName: &account.SessionName,

				MFASerialNumber: aws.String(""),
				MFATokenCode:    aws.String(""),
			})

			stsClient := sts.New(sess, conf)
			identity, err := stsClient.GetCallerIdentity(&sts.GetCallerIdentityInput{})
			if err != nil {
				return nil, err
			}

			result.Sessions = append(result.Sessions, &Session{
				Session:   sess,
				Config:    conf,
				AccountID: *identity.Account,
			})
		}
	}

	return result, nil
}
