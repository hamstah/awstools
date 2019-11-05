package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/hamstah/awstools/aws/dump/resources"
	"github.com/hamstah/awstools/common"
	log "github.com/sirupsen/logrus"

	kingpin "gopkg.in/alecthomas/kingpin.v2"
)

var (
	accountsConfigFilename         = kingpin.Flag("accounts-config", "Configuration file with the accounts to list resources for.").Short('c').String()
	terraformBackendConfigFilename = kingpin.Flag("terraform-backends-config", "Configuration file with the terraform backends to compare with.").Short('t').String()
	outputFilename                 = kingpin.Flag("output", "Filename to store the results in.").Short('o').String()
	onlyUnmanaged                  = kingpin.Flag("only-unmanaged", "Only return resources not managed by terraform.").Default("false").Bool()
	reports                        = kingpin.Flag("report", "Only run the specified report. Can be repeated.").Strings()
	listReports                    = kingpin.Flag("list-reports", "Prints the list of available reports and exits.").Default("false").Bool()
	startAsLambda                  = kingpin.Flag("start-as-lambda", "Start as lambda.").Default("false").Bool()
)

type Input struct {
	Accounts               []*resources.Account `json:"accounts"`
	TerraformBackendConfig *TerraformBackends   `json:"terraform_backend_config"`
	OnlyUnmanaged          bool                 `json:"only_unmanaged"`
	Reports                []string             `json:"reports"`
}

type Output struct {
	Resources []resources.Resource `json:"resources"`
}

func Handler() func(ctx context.Context, event Input) (*Output, error) {
	return func(ctx context.Context, event Input) (*Output, error) {
		output := &Output{}

		err := resources.OpenSessions(event.Accounts)
		if err != nil {
			return nil, err
		}

		services := resources.AllServices()

		jobs := []resources.Job{}

		if len(event.Reports) == 0 {
			for _, service := range services {
				for _, account := range event.Accounts {
					newJobs, err := service.GenerateAllJobs(account)
					common.FatalOnErrorW(err, "failed to generate jobs")
					jobs = append(jobs, newJobs...)
				}
			}
		} else {
			for _, name := range event.Reports {

				parts := strings.Split(name, ":")
				if len(parts) != 2 {
					common.Fatalln(fmt.Sprintf("Invalid report format %s, should be service:resource", name))
				}

				service, ok := services[parts[0]]
				if !ok {
					common.Fatalln(fmt.Sprintf("Invalid service %s", parts[0]))
				}

				for _, account := range event.Accounts {
					newJobs, err := service.GenerateJobs(account, parts[1])
					common.FatalOnErrorW(err, "failed to generate jobs")
					jobs = append(jobs, newJobs...)
				}
			}
		}

		result, errors := resources.Run(jobs)

		if event.TerraformBackendConfig != nil {

			err := event.TerraformBackendConfig.Pull()
			common.FatalOnErrorW(err, "failed to pull terraform state files")

			managed, err := event.TerraformBackendConfig.Load()
			common.FatalOnErrorW(err, "failed to load terraform state files")

			for _, resource := range result {
				s3Path, managed := managed[resource.UniqueID()]
				if managed {
					if event.OnlyUnmanaged {
						continue
					}
					resource.ManagedBy = map[string]string{
						"type":  "terraform",
						"state": s3Path,
					}
				}
				output.Resources = append(output.Resources, resource)
			}

		} else {
			output.Resources = result
		}

		for _, err := range errors {
			log.Error(err)
		}

		return output, nil
	}
}

func RunningInLambda() bool {
	// from https://docs.aws.amazon.com/lambda/latest/dg/lambda-environment-variables.html
	return strings.HasPrefix(os.Getenv("AWS_EXECUTION_ENV"), "AWS_Lambda_")
}

func main() {
	kingpin.CommandLine.Name = "aws-dump"
	kingpin.CommandLine.Help = "Dump AWS resources"
	common.HandleFlags()

	handler := Handler()

	if RunningInLambda() {
		lambda.Start(handler)
	} else {
		if *listReports {
			for _, report := range resources.AllReports() {
				fmt.Println(report)
			}
			os.Exit(0)
		}

		accounts, err := resources.NewAccountsFromFile(*accountsConfigFilename)
		common.FatalOnErrorW(err, "failed to load accounts from file")

		input := Input{
			Accounts:      accounts,
			Reports:       *reports,
			OnlyUnmanaged: *onlyUnmanaged,
		}

		if *terraformBackendConfigFilename != "" {
			backends, err := NewTerraformBackendsFromFile(*terraformBackendConfigFilename)
			common.FatalOnErrorW(err, "failed to load terraform backends from file")

			input.TerraformBackendConfig = backends
		}

		output, err := handler(context.Background(), input)
		common.FatalOnErrorW(err, "handler failed")

		reportJSON, err := json.MarshalIndent(output.Resources, "", "  ")
		common.FatalOnErrorW(err, "failed to serialise the report")

		err = ioutil.WriteFile(*outputFilename, reportJSON, 0644)
		common.FatalOnErrorW(err, "failed to write the report")
	}
}
