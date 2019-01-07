package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"

	"github.com/hamstah/awstools/common"

	kingpin "gopkg.in/alecthomas/kingpin.v2"
)

var (
	flags     = common.KingpinSessionFlags()
	infoFlags = common.KingpinInfoFlags()

	accountsConfig         = kingpin.Flag("accounts-config", "Configuration file with the accounts to list resources for.").Short('c').Required().String()
	terraformBackendConfig = kingpin.Flag("terraform-backends-config", "Configuration file with the terraform backends to compare with.").Short('t').String()
	output                 = kingpin.Flag("output", "Filename to store the results in.").Short('o').Required().String()
	onlyUnmanaged          = kingpin.Flag("only-unmanaged", "Only return resources not managed by terraform.").Default("false").Bool()
	reports                = kingpin.Flag("report", "Only run the specified report. Can be repeated.").Strings()
)

func main() {
	kingpin.CommandLine.Name = "aws-dump"
	kingpin.CommandLine.Help = "Dump AWS resources"
	kingpin.Parse()
	common.HandleInfoFlags(infoFlags)

	accounts, err := NewAccounts(*accountsConfig)
	common.FatalOnError(err)

	fetchers := map[string]Fetcher{
		"iam-list-groups":                IAMListGroups,
		"iam-list-users-and-access-keys": IAMListUsersAndAccessKeys,
		"iam-list-roles":                 IAMListRoles,
		"iam-list-policies":              IAMListPolicies,
		"s3-list-buckets":                S3ListBuckets,
		"ec2-list-security-groups":       EC2ListSecurityGroups,
		"ec2-list-vpcs":                  EC2ListVpcs,
		"cloudwatch-list-alarms":         CloudwatchListAlarms,
		"kms-list-aliases":               KMSListAliases,
		"kms-list-keys":                  KMSListKeys,
		"route53-list-hosted-zones":      Route53ListHostedZones,
	}

	enabledFetchers := []Fetcher{}
	if len(*reports) == 0 {
		for _, fetcher := range fetchers {
			enabledFetchers = append(enabledFetchers, fetcher)
		}
	} else {
		for _, name := range *reports {
			if fetcher, ok := fetchers[name]; ok {
				enabledFetchers = append(enabledFetchers, fetcher)
			} else {
				common.Fatalln(fmt.Sprintf("Invalid report %s", name))
			}
		}
	}

	resources := Run(accounts.Sessions, enabledFetchers)

	report := []Resource{}
	if *terraformBackendConfig != "" {
		backends, err := NewTerraformBackends(*terraformBackendConfig)
		common.FatalOnError(err)

		err = backends.Pull()
		common.FatalOnError(err)

		managed, err := backends.Load()
		common.FatalOnError(err)

		for _, resource := range resources {

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
		report = resources
	}

	reportJSON, err := json.MarshalIndent(report, "", "  ")
	common.FatalOnError(err)

	err = ioutil.WriteFile(*output, reportJSON, 0644)
	common.FatalOnError(err)
}
