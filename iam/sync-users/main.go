package main

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"os/user"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/hamstah/awstools/common"

	kingpin "gopkg.in/alecthomas/kingpin.v2"
)

var (
	flags     = common.KingpinSessionFlags()
	infoFlags = common.KingpinInfoFlags()
	groups    = kingpin.Flag("group", "Add users from this group. You can use --group multiple times.").Strings()
)

func main() {
	kingpin.CommandLine.Name = "iam-sync-users"
	kingpin.CommandLine.Help = "Create users from IAM"
	kingpin.Parse()
	common.HandleInfoFlags(infoFlags)
	common.FatalOnError(ensureCanCreateUser())

	session := session.Must(session.NewSession())
	conf := common.AssumeRoleConfig(flags, session)

	iamClient := iam.New(session, conf)

	users := []string{}

	for _, group := range *groups {
		newUsers, err := getUsersForGroup(iamClient, group)
		common.FatalOnError(err)
		users = append(users, newUsers...)
	}

	for _, userName := range users {
		_, err := user.Lookup(userName)
		if err == nil {
			// user already exists
			continue
		}

		common.FatalOnError(createUser(userName))
		fmt.Println(userName)
	}
}

func createUser(userName string) error {
	cmd := exec.Command("/usr/sbin/adduser", userName)
	err := cmd.Run()
	if err != nil {
		return err
	}

	sudoFilename := fmt.Sprintf("/etc/sudoers.d/%s", strings.Replace(userName, ".", "", -1))

	err = ioutil.WriteFile(sudoFilename, []byte(fmt.Sprintf("%s ALL=(ALL) NOPASSWD:ALL\n", userName)), 0644)
	if err != nil {
		return err
	}

	return nil
}

func ensureCanCreateUser() error {
	if _, err := os.Stat("/usr/sbin/adduser"); os.IsNotExist(err) {
		return errors.New("Can't find adduser to create user")
	}

	if _, err := os.Stat("/etc/sudoers.d"); os.IsNotExist(err) {
		return errors.New("Can't find adduser to create user")
	}
	return nil
}

func getUsersForGroup(client *iam.IAM, groupName string) ([]string, error) {
	users := []string{}
	err := client.GetGroupPages(&iam.GetGroupInput{GroupName: aws.String(groupName)},
		func(page *iam.GetGroupOutput, lastPage bool) bool {
			for _, user := range page.Users {
				users = append(users, *user.UserName)
			}
			return true
		})

	if err != nil {
		return nil, err
	}
	return users, nil
}
