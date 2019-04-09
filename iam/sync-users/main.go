package main

import (
	"bufio"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"os/user"
	"sort"
	"strconv"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/hamstah/awstools/common"
	log "github.com/sirupsen/logrus"
	kingpin "gopkg.in/alecthomas/kingpin.v2"
)

type IAMUser struct {
	Username string
	Groups   []string
	Sudo     bool
}

var (
	groups         = kingpin.Flag("group", "Add users from this IAM group. You can use --group multiple times.").Strings()
	iamTagsPrefix  = kingpin.Flag("iam-tags-prefix", "Prefix for tags in IAM").Default("iam-sync-users").String()
	lockMissing    = kingpin.Flag("lock-missing", "Lock local users not in IAM.").Default("false").Bool()
	lockIgnoreUser = kingpin.Flag("lock-ignore-user", "Ignore local user when locking.").Strings()
	sudo           = kingpin.Flag("sudo", "Add users to sudoers file.").Default("true").Bool()
)

func main() {
	kingpin.CommandLine.Name = "iam-sync-users"
	kingpin.CommandLine.Help = "Sync local users with IAM"
	flags := common.HandleFlags()
	common.FatalOnError(ensureCanCreateUser())

	session := session.Must(session.NewSession())
	conf := common.AssumeRoleConfig(flags, session)

	iamClient := iam.New(session, conf)

	users := []*IAMUser{}

	for _, group := range *groups {
		newUsers, err := getUsersForGroup(iamClient, group, *iamTagsPrefix)
		common.FatalOnError(err)
		users = append(users, newUsers...)
	}

	iamUsers := map[string]interface{}{}
	for _, u := range users {
		_, err := user.Lookup(u.Username)
		if err != nil {
			// user doesn't exists
			common.FatalOnError(createUser(u))

			_, err = user.Lookup(u.Username)
			common.FatalOnError(err)

		} else {
			err := UnlockLocalUser(u.Username)
			common.FatalOnError(err)
		}
		common.FatalOnError(syncUserSudo(u, *sudo))
		common.FatalOnError(syncUserGroups(u))
		iamUsers[u.Username] = 1
	}

	if *lockMissing {

		minimalUID, err := FindMinimalUID()
		common.FatalOnError(err)

		localUsers, err := LocalUsers()
		common.FatalOnError(err)

		ignoredLocalUsers := map[string]interface{}{}
		for _, ignoredLocalUser := range *lockIgnoreUser {
			ignoredLocalUsers[ignoredLocalUser] = 1
		}

		for _, localUser := range localUsers {
			if _, ok := iamUsers[localUser.Username]; ok {
				// exists in IAM, nothing to do
				continue
			}

			uid, err := strconv.Atoi(localUser.Uid)
			common.FatalOnError(err)

			if uid < minimalUID {
				// ignore system users
				continue
			}

			if _, ok := ignoredLocalUsers[localUser.Username]; ok {
				// do not lock local user
				continue
			}

			err = LockLocalUser(localUser.Username)
			common.FatalOnError(err)
		}
	}
}

func FindMinimalUID() (int, error) {
	file, err := os.Open("/etc/login.defs")
	if err != nil {
		return -1, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if len(line) == 0 || strings.HasPrefix(line, "#") {
			continue
		}

		parts := strings.Fields(line)
		if parts[0] != "UID_MIN" {
			continue
		}

		return strconv.Atoi(parts[len(parts)-1])
	}

	if err := scanner.Err(); err != nil {
		return -1, err
	}

	return -1, errors.New("Could not find UID_MIN")
}

func LocalUsers() ([]*user.User, error) {
	users := []*user.User{}

	file, err := os.Open("/etc/passwd")
	if err != nil {
		return nil, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		parts := strings.Split(scanner.Text(), ":")
		users = append(users, &user.User{
			Username: parts[0],
			Uid:      parts[2],
			Gid:      parts[3],
		})
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return users, err
}

func RunCommand(command string, args ...string) error {
	cmd := exec.Command(command, args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		log.WithFields(log.Fields{
			"exit_code": cmd.ProcessState.ExitCode(),
		}).Error(fmt.Sprintf("Failed to run command %s: %s", command, string(output)))
		return err
	}
	return nil
}

func LockLocalUser(username string) error {
	log.WithField("username", username).Info("Locking user")
	err := RunCommand("/usr/sbin/usermod", "-L", username)
	if err != nil {
		return err
	}
	return RunCommand("/usr/bin/chage", "-E", "-l", username)
}

func UnlockLocalUser(username string) error {
	log.WithField("username", username).Info("Unlocking user")
	err := RunCommand("/usr/sbin/usermod", "-U", username)
	if err != nil {
		return err
	}
	return RunCommand("/usr/bin/chage", "-E", "-1", username)
}

func syncUserGroups(iamUser *IAMUser) error {
	groupsStr := strings.Join(iamUser.Groups, ",")
	log.WithFields(log.Fields{
		"username": iamUser.Username,
		"groups": groupsStr,
		}).Info("Setting user groups")
	return RunCommand("/usr/sbin/usermod", "-G", groupsStr, iamUser.Username)
}

func syncUserSudo(iamUser *IAMUser, defaultSudo bool) error {
	sudoFilename := fmt.Sprintf("/etc/sudoers.d/%s", strings.Replace(iamUser.Username, ".", "", -1))
	withSudo := iamUser.Sudo || defaultSudo

	_, err := os.Stat(sudoFilename)
	hasSudo := (err == nil)

	if withSudo {
		// nothing to do
		if hasSudo {
			return nil
		}
		log.WithField("username", iamUser.Username).Info("Adding sudo to user")
		return ioutil.WriteFile(sudoFilename, []byte(fmt.Sprintf("%s ALL=(ALL) NOPASSWD:ALL\n", iamUser.Username)), 0644)
	} else {
		// nothing to do
		if !hasSudo {
			return nil
		}

		log.WithField("username", iamUser.Username).Info("Removing sudo from user")
		return os.Remove(sudoFilename)
	}
}

func createUser(iamUser *IAMUser) error {
	log.WithField("username", iamUser.Username).Info("Creating user")
	err := RunCommand("/usr/sbin/adduser", iamUser.Username)
	if err != nil {
		return err
	}

	return RunCommand("/usr/sbin/usermod", iamUser.Username, "-p", "'!'")
}

func ensureCanCreateUser() error {
	if _, err := os.Stat("/usr/sbin/adduser"); os.IsNotExist(err) {
		return errors.New("Can't find adduser to create user")
	}

	if _, err := os.Stat("/usr/sbin/usermod"); os.IsNotExist(err) {
		return errors.New("Can't find usermod to manage user groups")
	}

	if _, err := os.Stat("/usr/sbin/usermod"); os.IsNotExist(err) {
		return errors.New("Can't find usermod to manage user groups")
	}

	if _, err := os.Stat("/etc/login.defs"); os.IsNotExist(err) {
		return errors.New("Can't find /etc/login.defs to find minimal uid")
	}

	if _, err := os.Stat("/etc/passwd"); os.IsNotExist(err) {
		return errors.New("Can't find /etc/passwd to find local users")
	}

	if _, err := os.Stat("/etc/sudoers.d"); os.IsNotExist(err) {
		return errors.New("Can't find sudoers directory to create user")
	}
	return nil
}

func getUsersForGroup(client *iam.IAM, groupName string, iamTagsPrefix string) ([]*IAMUser, error) {
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

	output := make([]*IAMUser, 0, len(users))

	usersChan := make(chan string, len(users))
	results := make(chan *IAMUser, len(users))

	sudoTagName := fmt.Sprintf("%s:sudo", iamTagsPrefix)
	groupsTagName := fmt.Sprintf("%s:groups", iamTagsPrefix)

	for w := 0; w < 10; w++ {
		go func(usernames chan string, results chan *IAMUser) {
			for username := range usernames {
				res, err := client.ListUserTags(&iam.ListUserTagsInput{UserName: aws.String(username)})
				if err != nil {
					results <- nil
					continue
				}

				result := &IAMUser{
					Username: username,
					Groups:   []string{},
				}

				for _, tag := range res.Tags {

					switch *tag.Key {
					case groupsTagName:
						seen := map[string]interface{}{}
						for _, groupName := range strings.Split(*tag.Value, " ") {
							if groupName == "" {
								continue
							}
							if _, ok := seen[groupName]; ok {
								continue
							}
							seen[groupName] = true
							result.Groups = append(result.Groups, groupName)
						}
						sort.Strings(result.Groups)
					case sudoTagName:
						result.Sudo = *tag.Value == "true"
					}
				}

				results <- result
			}
		}(usersChan, results)
	}

	for _, username := range users {
		usersChan <- username
	}
	close(usersChan)

	for i := 0; i < len(users); i++ {
		result := <-results
		if result == nil {
			return nil, errors.New("Failed to list tags for user")
		}
		output = append(output, result)
	}

	return output, nil
}
