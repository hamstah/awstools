package resources

import (
	"strings"

	"github.com/aws/aws-sdk-go/service/rds"
	"github.com/fatih/structs"
)

var (
	RDSService = Service{
		Name: "rds",
		Reports: map[string]Report{
			"db-clusters":                   RDSListDBClusters,
			"db-instance-automated-backups": RDSListDBInstanceAutomatedBackups,
			"db-instances":                  RDSListDBInstances,
			"db-parameter-groups":           RDSListDBParameterGroups,
			"db-security-groups":            RDSListDBSecurityGroups,
			"db-snapshots":                  RDSListDBSnapshots,
			"db-subnet-groups":              RDSListDBSubnetGroups,
			"event-subscriptions":           RDSListEventSubscriptions,
			"events":                        RDSListEvents,
			"global-clusters":               RDSListGlobalClusters,
			"option-groups":                 RDSListOptionGroups,
			"reserved-db-instances":         RDSListReservedDBInstances,
		},
	}
)

func RDSListDBClusters(session *Session) *ReportResult {

	client := rds.New(session.Session, session.Config)

	resources := []Resource{}
	err := client.DescribeDBClustersPages(&rds.DescribeDBClustersInput{},
		func(page *rds.DescribeDBClustersOutput, lastPage bool) bool {
			for _, resource := range page.DBClusters {
				r := Resource{
					ID:        *resource.DBClusterIdentifier,
					ARN:       *resource.DBClusterArn,
					AccountID: session.AccountID,
					Service:   "rds",
					Type:      "db-cluster",
					Region:    *session.Config.Region,
					Metadata:  structs.Map(resource),
				}
				resources = append(resources, r)
			}

			return true
		})

	return &ReportResult{resources, err}
}

func RDSListDBInstanceAutomatedBackups(session *Session) *ReportResult {

	client := rds.New(session.Session, session.Config)

	resources := []Resource{}
	err := client.DescribeDBInstanceAutomatedBackupsPages(&rds.DescribeDBInstanceAutomatedBackupsInput{},
		func(page *rds.DescribeDBInstanceAutomatedBackupsOutput, lastPage bool) bool {
			for _, resource := range page.DBInstanceAutomatedBackups {
				r := Resource{
					ID:        *resource.DBInstanceArn,
					ARN:       *resource.DBInstanceArn,
					AccountID: session.AccountID,
					Service:   "rds",
					Type:      "db-instance-automated-backup",
					Region:    *session.Config.Region,
					Metadata:  structs.Map(resource),
				}
				resources = append(resources, r)
			}

			return true
		})

	return &ReportResult{resources, err}
}

func RDSListDBInstances(session *Session) *ReportResult {

	client := rds.New(session.Session, session.Config)

	resources := []Resource{}
	err := client.DescribeDBInstancesPages(&rds.DescribeDBInstancesInput{},
		func(page *rds.DescribeDBInstancesOutput, lastPage bool) bool {
			for _, resource := range page.DBInstances {
				r := Resource{
					ID:        *resource.DBInstanceIdentifier,
					ARN:       *resource.DBInstanceArn,
					AccountID: session.AccountID,
					Service:   "rds",
					Type:      "db-instance",
					Region:    *session.Config.Region,
					Metadata:  structs.Map(resource),
				}
				resources = append(resources, r)
			}

			return true
		})

	return &ReportResult{resources, err}
}

func RDSListDBParameterGroups(session *Session) *ReportResult {

	client := rds.New(session.Session, session.Config)

	resources := []Resource{}
	err := client.DescribeDBParameterGroupsPages(&rds.DescribeDBParameterGroupsInput{},
		func(page *rds.DescribeDBParameterGroupsOutput, lastPage bool) bool {
			for _, resource := range page.DBParameterGroups {
				r := Resource{
					ID:        *resource.DBParameterGroupArn,
					ARN:       *resource.DBParameterGroupArn,
					AccountID: session.AccountID,
					Service:   "rds",
					Type:      "db-parameter-group",
					Region:    *session.Config.Region,
					Metadata:  structs.Map(resource),
				}
				resources = append(resources, r)
			}

			return true
		})

	return &ReportResult{resources, err}
}

func RDSListDBSecurityGroups(session *Session) *ReportResult {

	client := rds.New(session.Session, session.Config)

	resources := []Resource{}
	err := client.DescribeDBSecurityGroupsPages(&rds.DescribeDBSecurityGroupsInput{},
		func(page *rds.DescribeDBSecurityGroupsOutput, lastPage bool) bool {
			for _, resource := range page.DBSecurityGroups {
				r := Resource{
					ID:        *resource.DBSecurityGroupArn,
					ARN:       *resource.DBSecurityGroupArn,
					AccountID: session.AccountID,
					Service:   "rds",
					Type:      "db-security-group",
					Region:    *session.Config.Region,
					Metadata:  structs.Map(resource),
				}
				resources = append(resources, r)
			}

			return true
		})

	return &ReportResult{resources, err}
}

func RDSListDBSnapshots(session *Session) *ReportResult {

	client := rds.New(session.Session, session.Config)

	resources := []Resource{}
	err := client.DescribeDBSnapshotsPages(&rds.DescribeDBSnapshotsInput{},
		func(page *rds.DescribeDBSnapshotsOutput, lastPage bool) bool {
			for _, resource := range page.DBSnapshots {
				r := Resource{
					ID:        *resource.DBSnapshotIdentifier,
					ARN:       *resource.DBSnapshotArn,
					AccountID: session.AccountID,
					Service:   "rds",
					Type:      "db-snapshot",
					Region:    *session.Config.Region,
					Metadata:  structs.Map(resource),
				}
				resources = append(resources, r)
			}

			return true
		})

	return &ReportResult{resources, err}
}

func RDSListDBSubnetGroups(session *Session) *ReportResult {

	client := rds.New(session.Session, session.Config)

	resources := []Resource{}
	err := client.DescribeDBSubnetGroupsPages(&rds.DescribeDBSubnetGroupsInput{},
		func(page *rds.DescribeDBSubnetGroupsOutput, lastPage bool) bool {
			for _, resource := range page.DBSubnetGroups {
				r := Resource{
					ID:        *resource.DBSubnetGroupArn,
					ARN:       *resource.DBSubnetGroupArn,
					AccountID: session.AccountID,
					Service:   "rds",
					Type:      "db-subnet-group",
					Region:    *session.Config.Region,
					Metadata:  structs.Map(resource),
				}
				resources = append(resources, r)
			}

			return true
		})

	return &ReportResult{resources, err}
}

func RDSListEventSubscriptions(session *Session) *ReportResult {

	client := rds.New(session.Session, session.Config)

	resources := []Resource{}
	err := client.DescribeEventSubscriptionsPages(&rds.DescribeEventSubscriptionsInput{},
		func(page *rds.DescribeEventSubscriptionsOutput, lastPage bool) bool {
			for _, resource := range page.EventSubscriptionsList {
				r := Resource{
					ID:        *resource.EventSubscriptionArn,
					ARN:       *resource.EventSubscriptionArn,
					AccountID: session.AccountID,
					Service:   "rds",
					Type:      "event-subscription",
					Region:    *session.Config.Region,
					Metadata:  structs.Map(resource),
				}
				resources = append(resources, r)
			}

			return true
		})

	return &ReportResult{resources, err}
}

func RDSListEvents(session *Session) *ReportResult {

	client := rds.New(session.Session, session.Config)

	resources := []Resource{}
	err := client.DescribeEventsPages(&rds.DescribeEventsInput{},
		func(page *rds.DescribeEventsOutput, lastPage bool) bool {
			for _, resource := range page.Events {
				r := Resource{
					ID:        *resource.SourceArn,
					ARN:       *resource.SourceArn,
					AccountID: session.AccountID,
					Service:   "rds",
					Type:      "event",
					Region:    *session.Config.Region,
					Metadata:  structs.Map(resource),
				}
				resources = append(resources, r)
			}

			return true
		})

	return &ReportResult{resources, err}
}

func RDSListGlobalClusters(session *Session) *ReportResult {

	client := rds.New(session.Session, session.Config)

	resources := []Resource{}
	err := client.DescribeGlobalClustersPages(&rds.DescribeGlobalClustersInput{},
		func(page *rds.DescribeGlobalClustersOutput, lastPage bool) bool {
			for _, resource := range page.GlobalClusters {
				r := Resource{
					ID:        *resource.GlobalClusterIdentifier,
					ARN:       *resource.GlobalClusterArn,
					AccountID: session.AccountID,
					Service:   "rds",
					Type:      "global-cluster",
					Region:    *session.Config.Region,
					Metadata:  structs.Map(resource),
				}
				resources = append(resources, r)
			}

			return true
		})

	return &ReportResult{resources, err}
}

func RDSListOptionGroups(session *Session) *ReportResult {

	client := rds.New(session.Session, session.Config)

	resources := []Resource{}
	err := client.DescribeOptionGroupsPages(&rds.DescribeOptionGroupsInput{},
		func(page *rds.DescribeOptionGroupsOutput, lastPage bool) bool {
			for _, resource := range page.OptionGroupsList {
				r := Resource{
					ID:        *resource.OptionGroupArn,
					ARN:       *resource.OptionGroupArn,
					AccountID: session.AccountID,
					Service:   "rds",
					Type:      "option-group",
					Region:    *session.Config.Region,
					Metadata:  structs.Map(resource),
				}
				if strings.HasPrefix(*resource.OptionGroupName, "default:") {
					continue
				}
				resources = append(resources, r)
			}

			return true
		})

	return &ReportResult{resources, err}
}

func RDSListReservedDBInstances(session *Session) *ReportResult {

	client := rds.New(session.Session, session.Config)

	resources := []Resource{}
	err := client.DescribeReservedDBInstancesPages(&rds.DescribeReservedDBInstancesInput{},
		func(page *rds.DescribeReservedDBInstancesOutput, lastPage bool) bool {
			for _, resource := range page.ReservedDBInstances {
				r := Resource{
					ID:        *resource.ReservedDBInstanceArn,
					ARN:       *resource.ReservedDBInstanceArn,
					AccountID: session.AccountID,
					Service:   "rds",
					Type:      "reserved-db-instance",
					Region:    *session.Config.Region,
					Metadata:  structs.Map(resource),
				}
				resources = append(resources, r)
			}

			return true
		})

	return &ReportResult{resources, err}
}
