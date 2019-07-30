package main

import (
	"github.com/aws/aws-sdk-go/service/autoscaling"
	"github.com/fatih/structs"
)

var (
	AutoScalingService = Service{
		Name: "autoscaling",
		Reports: map[string]Report{
			"groups": AutoScalingListGroups,
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
