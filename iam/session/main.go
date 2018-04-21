package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sts"
	"github.com/hamstah/awstools/common"
	kingpin "gopkg.in/alecthomas/kingpin.v2"
	ini "gopkg.in/ini.v1"
)

var (
	flags            = common.KingpinSessionFlags()
	quiet            = kingpin.Flag("quiet", "Do not output anything").Short('q').Default("false").Bool()
	saveProfileName  = kingpin.Flag("save-profile", "Save the profile in the AWS credentials storage").Short('s').String()
	overwriteProfile = kingpin.Flag("overwrite-profile", "Overwrite the profile if it already exists").Default("false").Bool()
	command          = kingpin.Arg("command", "Command to run, prefix with -- to pass args").Strings()
)

func main() {
	kingpin.CommandLine.Name = "iam-session"
	kingpin.CommandLine.Help = "Start a new session under a different role."
	kingpin.Parse()

	if len(*flags.RoleArn) == 0 && len(*saveProfileName) != 0 {
		common.Fatalln("--save-profile can only be used with --assume-role-arn")
	}

	if len(*command) == 0 && len(*saveProfileName) == 0 {
		common.Fatalln("Use at least one of command or --save-profile-name")
	}

	session := session.Must(session.NewSession())
	conf := common.AssumeRoleConfig(flags, session)

	stsClient := sts.New(session, conf)
	res, err := stsClient.GetCallerIdentity(&sts.GetCallerIdentityInput{})
	common.FatalOnError(err)

	if !*quiet {
		fmt.Println(res)
	}

	var creds credentials.Value
	if conf.Credentials != nil {
		creds, err = conf.Credentials.Get()
		common.FatalOnError(err)
	}

	if len(*saveProfileName) != 0 {
		saveProfile(conf, &creds)
	}

	if len(*command) > 0 {
		executeCommand(command, conf, &creds)
	}
}

func saveProfile(conf *aws.Config, creds *credentials.Value) {
	// update the credentials file
	credsFilename := os.ExpandEnv("$HOME/.aws/credentials")
	credsCfg, err := ini.Load(credsFilename)
	common.FatalOnError(err)

	_, err = credsCfg.GetSection(*saveProfileName)
	if err == nil {
		// section already exists, prompt to delete or not unless the override flag is used
		if !*overwriteProfile {
			confirm := promptConfirm(fmt.Sprintf("The profile %s already exists, do you want to override it? (y/n) [n]: ", *saveProfileName))
			if !confirm {
				fmt.Println("Not overwriting profile")
				os.Exit(0)
			}
		}
		if !*quiet {
			fmt.Println("Overwriting", *saveProfileName)
		}
		credsCfg.DeleteSection(*saveProfileName)
	}

	newCredsSection, err := credsCfg.NewSection(*saveProfileName)
	common.FatalOnError(err)
	_, err = newCredsSection.NewKey("aws_access_key_id", creds.AccessKeyID)
	common.FatalOnError(err)
	_, err = newCredsSection.NewKey("aws_secret_access_key", creds.SecretAccessKey)
	common.FatalOnError(err)
	_, err = newCredsSection.NewKey("aws_session_token", creds.SessionToken)
	common.FatalOnError(err)

	// update the config file
	configFilename := os.ExpandEnv("$HOME/.aws/config")
	configCfg, err := ini.Load(configFilename)
	common.FatalOnError(err)

	configSectionName := fmt.Sprintf("profile %s", *saveProfileName)
	_, err = configCfg.GetSection(configSectionName)
	if err == nil {
		configCfg.DeleteSection(configSectionName)
	}

	newConfigSection, err := configCfg.NewSection(configSectionName)
	common.FatalOnError(err)
	_, err = newConfigSection.NewKey("region", *conf.Region)
	common.FatalOnError(err)
	_, err = newConfigSection.NewKey("format", "json")
	common.FatalOnError(err)

	credsCfg.SaveTo(credsFilename)
	configCfg.SaveTo(configFilename)
}

func promptConfirm(text string) bool {
	var response string
	fmt.Print(text)
	_, err := fmt.Scanln(&response)
	common.FatalOnError(err)
	fmt.Println()
	return response == "y"
}

func executeCommand(command *[]string, conf *aws.Config, creds *credentials.Value) {
	env := os.Environ()
	var pEnv []string
	if conf.Credentials != nil {
		pEnv = []string{
			fmt.Sprintf("AWS_ACCESS_KEY_ID=%s", creds.AccessKeyID),
			fmt.Sprintf("AWS_SECRET_ACCESS_KEY=%s", creds.SecretAccessKey),
			fmt.Sprintf("AWS_SESSION_TOKEN=%s", creds.SessionToken),
			fmt.Sprintf("AWS_REGION=%s", *conf.Region),
		}
		for _, v := range env {
			s := strings.SplitN(v, "=", 2)
			if strings.HasPrefix(s[0], "AWS") {
				continue
			}
			pEnv = append(pEnv, v)
		}
	} else {
		pEnv = env
	}
	if !*quiet {
		fmt.Println("running", *command)
	}
	p := exec.Command((*command)[0], (*command)[1:]...)
	p.Env = pEnv
	p.Stdin = os.Stdin
	p.Stderr = os.Stderr
	p.Stdout = os.Stdout
	p.Run()
}
