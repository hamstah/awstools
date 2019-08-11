package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/hamstah/awstools/aws/dump/resources"
	"github.com/hamstah/awstools/common"
	log "github.com/sirupsen/logrus"

	kingpin "gopkg.in/alecthomas/kingpin.v2"
)

var (
	accountsConfig         = kingpin.Flag("accounts-config", "Configuration file with the accounts to list resources for.").Short('c').Required().String()
	terraformBackendConfig = kingpin.Flag("terraform-backends-config", "Configuration file with the terraform backends to compare with.").Short('t').String()
	output                 = kingpin.Flag("output", "Filename to store the results in.").Short('o').Required().String()
	onlyUnmanaged          = kingpin.Flag("only-unmanaged", "Only return resources not managed by terraform.").Default("false").Bool()
	reports                = kingpin.Flag("report", "Only run the specified report. Can be repeated.").Strings()
	listReports            = kingpin.Flag("list-reports", "Prints the list of available reports and exits.").Default("false").Bool()
)

func main() {
	kingpin.CommandLine.Name = "aws-dump"
	kingpin.CommandLine.Help = "Dump AWS resources"
	common.HandleFlags()

	if *listReports {
		for _, report := range resources.AllReports() {
			fmt.Println(report)
		}
		os.Exit(0)
	}

	accounts, err := resources.NewAccounts(*accountsConfig)
	common.FatalOnError(err)

	services := resources.AllServices()

	jobs := []resources.Job{}

	if len(*reports) == 0 {
		for _, service := range services {
			for _, account := range accounts.Accounts {
				newJobs, err := service.GenerateAllJobs(account)
				common.FatalOnError(err)
				jobs = append(jobs, newJobs...)
			}
		}
	} else {
		for _, name := range *reports {

			parts := strings.Split(name, ":")
			if len(parts) != 2 {
				common.Fatalln(fmt.Sprintf("Invalid report format %s, should be service:resource", name))
			}

			service, ok := services[parts[0]]
			if !ok {
				common.Fatalln(fmt.Sprintf("Invalid service %s", parts[0]))
			}

			for _, account := range accounts.Accounts {
				newJobs, err := service.GenerateJobs(account, parts[1])
				common.FatalOnError(err)
				jobs = append(jobs, newJobs...)
			}
		}
	}

	result, errors := resources.Run(jobs)

	report := []resources.Resource{}
	if *terraformBackendConfig != "" {
		backends, err := NewTerraformBackends(*terraformBackendConfig)
		common.FatalOnError(err)

		err = backends.Pull()
		common.FatalOnError(err)

		managed, err := backends.Load()
		common.FatalOnError(err)

		for _, resource := range result {

			s3Path, managed := managed[resource.UniqueID()]
			if managed {
				if *onlyUnmanaged {
					continue
				}
				resource.ManagedBy = map[string]string{
					"type":  "terraform",
					"state": s3Path,
				}
			}
			report = append(report, resource)
		}

	} else {
		report = result
	}

	for _, err := range errors {
		log.Error(err)
	}

	reportJSON, err := json.MarshalIndent(report, "", "  ")
	common.FatalOnError(err)

	err = ioutil.WriteFile(*output, reportJSON, 0644)
	common.FatalOnError(err)
}
