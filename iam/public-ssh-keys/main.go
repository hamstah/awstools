package main

import (
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/aws/aws-sdk-go/service/sts"
	"github.com/hamstah/awstools/common"

	kingpin "gopkg.in/alecthomas/kingpin.v2"
)

var (
	flags    = common.KingpinSessionFlags()
	username = kingpin.Flag("username", "Username to fetch the keys for, otherwise default to the logged in user.").Short('u').String()
	encoding = kingpin.Flag("key-encoding", "Encoding of the key to return (SSH or PEM)").Default("SSH").Enum("PEM", "SSH")
	groups   = kingpin.Flag("allowed-group", "Fetch the keys only if the user is in this group. You can use --allowed-group multiple times.").Strings()
)

func main() {
	kingpin.CommandLine.Name = "iam-public-ssh-keys"
	kingpin.CommandLine.Help = "Return public SSH keys for an IAM user."
	kingpin.Parse()

	session := session.Must(session.NewSession())
	conf := common.AssumeRoleConfig(flags, session)

	iamClient := iam.New(session, conf)

	if len(*username) == 0 {
		stsClient := sts.New(session, conf)
		res, err := stsClient.GetCallerIdentity(&sts.GetCallerIdentityInput{})
		common.FatalOnError(err)

		p := strings.Split(*res.Arn, "/")
		username = aws.String(p[len(p)-1])
	}

	if len(*groups) != 0 {
		found, err := checkGroups(iamClient, username, groups)
		common.FatalOnError(err)
		if !found {
			common.Fatalln("User is not part of the allowed group")
		}
	}

	sshKeys, err := getSSHKeys(iamClient, username)
	common.FatalOnError(err)

	for _, sshKey := range sshKeys {
		fmt.Println(*sshKey)
	}
}

func getSSHKeys(iamClient *iam.IAM, username *string) ([]*string, error) {
	sshParams := &iam.ListSSHPublicKeysInput{
		UserName: username,
	}

	sshKeyIDs := make(chan *string, 100)
	sshKeys := make(chan *string, 100)

	for w := 0; w < 5; w++ {
		go worker(iamClient, encoding, username, sshKeyIDs, sshKeys)
	}

	maxPage := 10
	pageNum := 0
	total := 0
	err := iamClient.ListSSHPublicKeysPages(sshParams,
		func(page *iam.ListSSHPublicKeysOutput, lastPage bool) bool {
			pageNum++
			for _, sshKey := range page.SSHPublicKeys {
				if *sshKey.Status == "Active" {
					sshKeyIDs <- sshKey.SSHPublicKeyId
					total++
				}
			}
			return pageNum <= maxPage
		})

	if err != nil {
		return nil, err
	}

	result := make([]*string, 0, total)
	for a := 0; a < total; a++ {
		res := <-sshKeys
		if res != nil {
			result = append(result, res)
		}
	}
	return result, nil
}

func checkGroups(iamClient *iam.IAM, username *string, groups *[]string) (bool, error) {
	maxPage := 10

	lookup := make(map[string]bool, len(*groups))
	for _, group := range *groups {
		lookup[group] = true
	}
	groupsParams := &iam.ListGroupsForUserInput{
		UserName: username,
	}
	pageNum := 0
	found := false
	err := iamClient.ListGroupsForUserPages(groupsParams,
		func(page *iam.ListGroupsForUserOutput, lastPage bool) bool {
			pageNum++
			for _, group := range page.Groups {
				if lookup[*group.GroupName] {
					found = true
					return false
				}
			}
			return pageNum <= maxPage
		})
	if err != nil {
		return false, err
	}
	return found, nil
}

func worker(iamClient *iam.IAM, encoding, username *string, sshKeysIDs <-chan *string, sshKeys chan<- *string) {
	for sshKeyID := range sshKeysIDs {
		result, err := iamClient.GetSSHPublicKey(&iam.GetSSHPublicKeyInput{
			SSHPublicKeyId: sshKeyID,
			Encoding:       encoding,
			UserName:       username,
		})
		if err != nil {
			common.ErrorLog.Println(err)
			sshKeys <- nil
		} else {
			sshKeys <- result.SSHPublicKey.SSHPublicKeyBody
		}
	}
}
