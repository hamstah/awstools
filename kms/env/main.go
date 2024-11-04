package main

import (
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"reflect"
	"syscall"
	"time"

	kingpin "github.com/alecthomas/kingpin/v2"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/hamstah/awstools/common"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

var (
	command                    = kingpin.Arg("command", "Command to run, prefix with -- to pass args").Required().Strings()
	kmsPrefix                  = kingpin.Flag("kms-prefix", "Prefix for the KMS environment variables").Default("KMS_").String()
	ssmPrefix                  = kingpin.Flag("ssm-prefix", "Prefix for the SSM environment variables").Default("SSM_").String()
	secretsManagerPrefix       = kingpin.Flag("secrets-manager-prefix", "Prefix for the secrets manager environment variables").Default("SECRETS_MANAGER_").String()
	secretsManagerVersionStage = kingpin.Flag("secrets-manager-version-stage", "The version stage of secrets from secrets manager").Default("AWSCURRENT").String()
	refreshInterval            = kingpin.Flag("refresh-interval", "Refresh interval").Default("0").Duration()
	refreshAction              = kingpin.Flag("refresh-action", "Action to take when values have changed").Default("RESTART").Enum("RESTART", "EXIT")
	refreshMaxRetries          = kingpin.Flag("refresh-max-retries", "Number of retries when failing to refresh the config").Default("5").Int()
)

func RefreshAndFlatten(config *common.ConfigValues, session *session.Session, awsConfig *aws.Config) (map[string]string, error) {

	value := map[string]interface{}{}

	err := config.RefreshWithRetries(session, awsConfig, &value)
	if err != nil {
		return nil, errors.Wrap(err, "failed to refresh config")
	}

	flat, err := common.FlattenEnvVarMap(value)
	if err != nil {
		return nil, errors.Wrap(err, "failed to flatten config map, make sure it doesn't contain slices")
	}

	return flat, nil
}

func Monitor(flags *common.SessionFlags, config *common.ConfigValues, comm chan<- map[string]string) {

	session, awsConfig := common.OpenSession(flags)

	previous, err := RefreshAndFlatten(config, session, awsConfig)
	if err != nil {
		log.Error(err)
		comm <- nil
		return
	}
	comm <- previous

	if !config.IsRefreshable() || *refreshInterval == time.Duration(0) {
		return
	}

	for range time.Tick(*refreshInterval) {
		new, err := RefreshAndFlatten(config, session, awsConfig)
		if err != nil {
			log.Error(err)
			comm <- nil
			return
		}

		if !reflect.DeepEqual(new, previous) {
			comm <- new
		}
		previous = new
	}
}

func main() {
	kingpin.CommandLine.Name = "kms-env"
	kingpin.CommandLine.Help = "Decrypt environment variables encrypted with KMS, SSM or Secret Manager."
	flags := common.HandleFlags()

	config := common.NewConfigValues()
	config.MaxRetries = *refreshMaxRetries
	config.KeyPrefixes = map[string]string{
		"KMS":             *kmsPrefix,
		"SSM":             *ssmPrefix,
		"SECRETS_MANAGER": *secretsManagerPrefix,
		"FILE":            "FILE_",
	}
	config.Settings["secrets_manager_version_stage"] = *secretsManagerVersionStage

	err := config.SetFromEnvironment()
	common.FatalOnError(err)

	comm := make(chan map[string]string, 1)
	go Monitor(flags, config, comm)

	var p *exec.Cmd
	waitingPid := -1

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	for envMap := range comm {
		if envMap == nil {
			// failed to refresh config, don't kill the process
			// stay alive even if config potentially out of date is better than a crash
			continue
		}

		if p != nil {
			waitingPid = p.Process.Pid
			_ = p.Process.Signal(syscall.SIGTERM)
			_ = p.Wait()

			if *refreshAction == "EXIT" {
				os.Exit(0)
			}
		}

		var pEnv []string
		for key, value := range envMap {
			pEnv = append(pEnv, fmt.Sprintf("%s=%s", key, value))
		}
		p = exec.Command((*command)[0], (*command)[1:]...)
		p.Env = pEnv
		p.Stdin = os.Stdin
		p.Stderr = os.Stderr
		p.Stdout = os.Stdout
		err := p.Start()
		common.FatalOnError(err)

		// forward signals to the child process
		go func(p *exec.Cmd) {
			for sig := range sigChan {
				if p.Process != nil {
					p.Process.Signal(sig)
				}
			}
		}(p)

		go func(p *exec.Cmd) {
			p.Wait()
			if waitingPid != p.Process.Pid {
				os.Exit(p.ProcessState.ExitCode())
			}
		}(p)
	}
}
