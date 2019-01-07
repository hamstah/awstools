package main

import (
	"fmt"

	"github.com/aws/aws-sdk-go/service/ec2"
)

func EC2ListVpcs(session *Session) *FetchResult {
	client := ec2.New(session.Session, session.Config)

	vpcs := []Resource{}

	res, err := client.DescribeVpcs(&ec2.DescribeVpcsInput{})
	if err != nil {
		return &FetchResult{nil, err}
	}

	for _, vpc := range res.Vpcs {
		if *vpc.IsDefault {
			continue
		}
		vpcs = append(vpcs, Resource{
			ID: *vpc.VpcId,
			// ARN
			Service:   "ec2",
			Type:      "vpc",
			AccountID: *vpc.OwnerId,
			Region:    *session.Config.Region,
		})
	}

	return &FetchResult{vpcs, err}
}

func EC2ListSecurityGroups(session *Session) *FetchResult {
	client := ec2.New(session.Session, session.Config)

	securityGroups := []Resource{}
	err := client.DescribeSecurityGroupsPages(&ec2.DescribeSecurityGroupsInput{},
		func(page *ec2.DescribeSecurityGroupsOutput, lastPage bool) bool {
			for _, securityGroup := range page.SecurityGroups {
				resource := Resource{
					ID: *securityGroup.GroupId,
					ARN: fmt.Sprintf("arn:aws:ec2:%s:%s:security-group/%s",
						*session.Config.Region,
						*securityGroup.OwnerId,
						*securityGroup.GroupId,
					),
					Service:   "ec2",
					Type:      "security-group",
					AccountID: *securityGroup.OwnerId,
					Region:    *session.Config.Region,
					Metadata: map[string]string{
						"GroupName":   *securityGroup.GroupName,
						"Description": *securityGroup.Description,
					},
				}
				if securityGroup.VpcId != nil {
					resource.Metadata["VpcId"] = *securityGroup.VpcId
				}
				securityGroups = append(securityGroups, resource)
			}

			return true
		})

	return &FetchResult{securityGroups, err}
}
