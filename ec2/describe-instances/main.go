package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	kingpin "github.com/alecthomas/kingpin/v2"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hamstah/awstools/common"
)

var (
	filter      = kingpin.Flag("filter", "The filter to use for the identifiers. eg tag:Name").String()
	identifiers = kingpin.Arg("identifiers", "If omitted the instance is fetched from the EC2 metadata.").Strings()
)

type SecurityGroup struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type Result struct {
	InstanceID         string            `json:"instance_id"`
	IAMInstanceProfile string            `json:"iam_instance_profile"`
	PrivateDNSName     string            `json:"private_dns_name"`
	PrivateIPAddress   string            `json:"private_ip_address"`
	PublicDNSName      string            `json:"public_dns_name"`
	PublicIPAddress    string            `json:"public_ip_address"`
	ImageID            string            `json:"image_id"`
	InstanceType       string            `json:"instance_type"`
	KeyName            string            `json:"key_name"`
	LaunchTime         time.Time         `json:"launch_time"`
	SubnetID           string            `json:"subnet_id"`
	Tags               map[string]string `json:"tags"`
	SecurityGroups     []*SecurityGroup  `json:"security_groups"`
	State              string            `json:"state"`
	VPCID              string            `json:"vpc_id"`
	AutoScalingGroup   string            `json:"auto_scaling_group"`
}

func main() {
	kingpin.CommandLine.Name = "ec2-describe-instances"
	kingpin.CommandLine.Help = "Returns metadata of one or more EC2 instances"
	flags := common.HandleFlags()

	session, conf := common.OpenSession(flags)

	client := ec2.New(session, conf)

	input := &ec2.DescribeInstancesInput{}

	if len(*identifiers) == 0 {
		resp, err := http.Get("http://169.254.169.254/latest/meta-data/instance-id")
		common.FatalOnError(err)
		defer resp.Body.Close()

		body, err := ioutil.ReadAll(resp.Body)
		common.FatalOnError(err)

		input.InstanceIds = []*string{aws.String(string(body))}
	} else {
		values := []*string{}
		for _, identifier := range *identifiers {
			values = append(values, aws.String(identifier))
		}

		if *filter == "" {
			input.InstanceIds = values
		} else {
			input.Filters = []*ec2.Filter{
				{
					Name:   filter,
					Values: values,
				},
			}
		}
	}

	results := []*Result{}
	err := client.DescribeInstancesPages(input,
		func(page *ec2.DescribeInstancesOutput, lastPage bool) bool {
			for _, reservation := range page.Reservations {
				result := DescribeInstance(reservation)
				if result != nil {
					results = append(results, result)
				}
			}

			return true
		})
	common.FatalOnError(err)

	data, err := json.MarshalIndent(results, "", "  ")
	common.FatalOnError(err)

	fmt.Println(string(data))
}

func MaybeString(value *string) string {
	if value != nil {
		return *value
	}
	return ""
}

func DescribeInstance(reservation *ec2.Reservation) *Result {
	result := &Result{
		Tags:           map[string]string{},
		SecurityGroups: []*SecurityGroup{},
	}

	if len(reservation.Instances) > 0 {
		for _, instance := range reservation.Instances {
			if instance == nil {
				return nil
			}

			result.InstanceID = *instance.InstanceId

			if instance.PrivateIpAddress != nil {
				result.PrivateIPAddress = *instance.PrivateIpAddress
				result.PrivateDNSName = *instance.PrivateDnsName
			}

			if instance.PublicIpAddress != nil {
				result.PublicIPAddress = *instance.PublicIpAddress
				result.PublicDNSName = *instance.PublicDnsName
			}
			if instance.IamInstanceProfile != nil {
				result.IAMInstanceProfile = *instance.IamInstanceProfile.Arn
			}

			result.LaunchTime = *instance.LaunchTime
			result.InstanceType = *instance.InstanceType
			result.ImageID = *instance.ImageId
			result.KeyName = *instance.KeyName

			result.SubnetID = MaybeString(instance.SubnetId)
			result.VPCID = MaybeString(instance.VpcId)
			result.State = *instance.State.Name

			for _, securityGroup := range instance.SecurityGroups {
				result.SecurityGroups = append(result.SecurityGroups, &SecurityGroup{
					ID:   *securityGroup.GroupId,
					Name: *securityGroup.GroupName,
				})
			}

			for _, tag := range instance.Tags {
				result.Tags[*tag.Key] = *tag.Value
				if *tag.Key == "aws:autoscaling:groupName" {
					result.AutoScalingGroup = *tag.Value
				}
			}
		}
	}
	return result
}
