package resources

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
	RolePolicy  string   `json:"role_policy"`
	ExternalID  string   `json:"external_id"`
	SessionName string   `json:"session_name"`
	Sessions    []*Session
}

type Session struct {
	Session   *session.Session
	Config    *aws.Config
	AccountID string
}

func NewAccountsFromFile(filename string) ([]*Account, error) {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	accounts := []*Account{}
	err = json.Unmarshal(data, &accounts)
	if err != nil {
		return nil, err
	}

	return accounts, nil
}

func OpenSessions(accounts []*Account) error {
	for _, account := range accounts {
		account.Sessions = []*Session{}
		for _, region := range account.Regions {
			sess, conf := common.OpenSession(&common.SessionFlags{
				RoleArn:         &account.RoleARN,
				RoleExternalID:  &account.ExternalID,
				RolePolicy:      &account.RolePolicy,
				Region:          &region,
				RoleSessionName: &account.SessionName,

				MFASerialNumber: aws.String(""),
				MFATokenCode:    aws.String(""),
			})

			stsClient := sts.New(sess, conf)
			identity, err := stsClient.GetCallerIdentity(&sts.GetCallerIdentityInput{})
			if err != nil {
				return err
			}
			session := &Session{
				Session:   sess,
				Config:    conf,
				AccountID: *identity.Account,
			}
			account.Sessions = append(account.Sessions, session)
		}
	}
	return nil
}
