package main

import (
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/fatih/structs"
)

var (
	EC2Service = Service{
		Name: "ec2",
		Reports: map[string]Report{
			"vpcs":            EC2ListVpcs,
			"security-groups": EC2ListSecurityGroups,
			"images":          EC2ListImages,
			"instances":       EC2ListInstances,
			"nat-gateways":    EC2ListNATGateways,
			"key-pairs":       EC2ListKeyPairs,
		},
	}
)

func EC2ListVpcs(session *Session) *ReportResult {
	client := ec2.New(session.Session, session.Config)

	vpcs := []Resource{}

	res, err := client.DescribeVpcs(&ec2.DescribeVpcsInput{})
	if err != nil {
		return &ReportResult{nil, err}
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
			Metadata:  structs.Map(vpc),
		})
	}

	return &ReportResult{vpcs, err}
}

func EC2ListSecurityGroups(session *Session) *ReportResult {
	client := ec2.New(session.Session, session.Config)
	result := &ReportResult{}
	groupIds := []*string{}
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
					Metadata:  structs.Map(securityGroup),
				}
				if securityGroup.VpcId != nil {
					resource.Metadata["VpcId"] = *securityGroup.VpcId
				}
				groupIds = append(groupIds, securityGroup.GroupId)
				result.Resources = append(result.Resources, resource)
			}

			return true
		})

	if err != nil {
		result.Error = err
		return result
	}

	// Max filter size is 200 values
	batchSize := 200
	var batches [][]*string

	for batchSize < len(groupIds) {
		groupIds, batches = groupIds[batchSize:], append(batches, groupIds[0:batchSize:batchSize])
	}
	batches = append(batches, groupIds)

	used := map[string]interface{}{}
	for _, batch := range batches {
		err := client.DescribeNetworkInterfacesPages(&ec2.DescribeNetworkInterfacesInput{
			Filters: []*ec2.Filter{
				&ec2.Filter{
					Name:   aws.String("group-id"),
					Values: batch,
				},
			},
		},
			func(page *ec2.DescribeNetworkInterfacesOutput, lastPage bool) bool {
				for _, networkInterface := range page.NetworkInterfaces {
					for _, group := range networkInterface.Groups {
						used[*group.GroupId] = 1
					}
				}
				return true
			})
		if err != nil {
			result.Error = err
			return result
		}
	}

	now := time.Now().UTC()
	for _, resource := range result.Resources {
		var lastUsed *time.Time
		if _, ok := used[resource.ID]; ok {
			lastUsed = &now
		}
		resource.Metadata["LastUsed"] = lastUsed
	}

	return result
}

func EC2ListImages(session *Session) *ReportResult {
	client := ec2.New(session.Session, session.Config)

	images := []Resource{}

	res, err := client.DescribeImages(&ec2.DescribeImagesInput{
		Owners: []*string{aws.String("self")},
	})
	if err != nil {
		return &ReportResult{nil, err}
	}

	for _, image := range res.Images {
		images = append(images, Resource{
			ID:        *image.ImageId,
			Service:   "ec2",
			Type:      "image",
			AccountID: *image.OwnerId,
			Region:    *session.Config.Region,
			Metadata:  structs.Map(image),
		})
	}

	return &ReportResult{images, err}
}

func EC2ListInstances(session *Session) *ReportResult {
	client := ec2.New(session.Session, session.Config)

	instances := []Resource{}
	err := client.DescribeInstancesPages(&ec2.DescribeInstancesInput{},
		func(page *ec2.DescribeInstancesOutput, lastPage bool) bool {
			for _, reservation := range page.Reservations {
				for _, instance := range reservation.Instances {
					resource := Resource{
						ID:        *instance.InstanceId,
						ARN:       "",
						AccountID: session.AccountID,
						Service:   "ec2",
						Type:      "instance",
						Region:    *session.Config.Region,
						Metadata:  structs.Map(instance),
					}
					instances = append(instances, resource)
				}
			}

			return true
		})

	return &ReportResult{instances, err}
}

func EC2ListNATGateways(session *Session) *ReportResult {

	client := ec2.New(session.Session, session.Config)

	resources := []Resource{}
	err := client.DescribeNatGatewaysPages(&ec2.DescribeNatGatewaysInput{},
		func(page *ec2.DescribeNatGatewaysOutput, lastPage bool) bool {
			for _, natGateway := range page.NatGateways {
				resource := Resource{
					ID:        *natGateway.NatGatewayId,
					ARN:       "",
					AccountID: session.AccountID,
					Service:   "ec2",
					Type:      "nat-gateway",
					Region:    *session.Config.Region,
					Metadata:  structs.Map(natGateway),
				}
				resources = append(resources, resource)
			}

			return true
		})

	return &ReportResult{resources, err}
}

func EC2ListKeyPairs(session *Session) *ReportResult {
	client := ec2.New(session.Session, session.Config)

	keypairs := []Resource{}

	res, err := client.DescribeKeyPairs(&ec2.DescribeKeyPairsInput{})
	if err != nil {
		return &ReportResult{nil, err}
	}

	for _, keypair := range res.KeyPairs {
		keypairs = append(keypairs, Resource{
			ID:        *keypair.KeyName,
			Service:   "ec2",
			Type:      "key-pair",
			AccountID: session.AccountID,
			Region:    *session.Config.Region,
			Metadata:  structs.Map(keypair),
		})
	}

	return &ReportResult{keypairs, err}
}
