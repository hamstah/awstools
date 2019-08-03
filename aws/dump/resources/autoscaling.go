package resources

import (
	"github.com/aws/aws-sdk-go/service/autoscaling"
	"github.com/fatih/structs"
)

var (
	AutoScalingService = Service{
		Name: "autoscaling",
		Reports: map[string]Report{
			"groups":                AutoScalingListGroups,
			"launch-configurations": AutoScalingListLaunchConfigurations,
		},
	}
)

func AutoScalingListGroups(session *Session) *ReportResult {

	client := autoscaling.New(session.Session, session.Config)

	resources := []Resource{}
	err := client.DescribeAutoScalingGroupsPages(&autoscaling.DescribeAutoScalingGroupsInput{},
		func(page *autoscaling.DescribeAutoScalingGroupsOutput, lastPage bool) bool {
			for _, autoScalingGroup := range page.AutoScalingGroups {
				resource := Resource{
					ID:        *autoScalingGroup.AutoScalingGroupName,
					ARN:       *autoScalingGroup.AutoScalingGroupARN,
					AccountID: session.AccountID,
					Service:   "autoscaling",
					Type:      "group",
					Region:    *session.Config.Region,
					Metadata:  structs.Map(autoScalingGroup),
				}
				resources = append(resources, resource)
			}

			return true
		})

	return &ReportResult{resources, err}
}

func AutoScalingListLaunchConfigurations(session *Session) *ReportResult {

	client := autoscaling.New(session.Session, session.Config)

	resources := []Resource{}
	err := client.DescribeLaunchConfigurationsPages(&autoscaling.DescribeLaunchConfigurationsInput{},
		func(page *autoscaling.DescribeLaunchConfigurationsOutput, lastPage bool) bool {
			for _, launchConfiguration := range page.LaunchConfigurations {
				resource := Resource{
					ID:        *launchConfiguration.LaunchConfigurationName,
					ARN:       *launchConfiguration.LaunchConfigurationARN,
					AccountID: session.AccountID,
					Service:   "autoscaling",
					Type:      "launch-configuration",
					Region:    *session.Config.Region,
					Metadata:  structs.Map(launchConfiguration),
				}
				resources = append(resources, resource)
			}

			return true
		})

	return &ReportResult{resources, err}
}
