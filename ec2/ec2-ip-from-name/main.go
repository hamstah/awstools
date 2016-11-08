package main

import (
	"flag"
	"fmt"
	"os"
	"sort"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
)

func main() {

	instanceName := flag.String("name", "", "Name of the EC2 instance")
	maxResults := flag.Int("max-results", 1, "Number of results")

	region := flag.String("region", "eu-west-1", "AWS region")
	flag.Parse()

	if len(*instanceName) < 1 {
		fmt.Println("Missing instance name")
		os.Exit(1)
	}

	config := aws.Config{Region: aws.String(*region)}
	session := session.New(&config)

	svc := ec2.New(session)
	params := &ec2.DescribeInstancesInput{
		Filters: []*ec2.Filter{
			{
				Name: aws.String("tag:Name"),
				Values: []*string{
					aws.String(*instanceName),
				},
			},
		},
	}
	resp, err := svc.DescribeInstances(params)

	if err != nil {
		fmt.Println(err.Error())
		os.Exit(3)
	}
	ips := make([]string, len(resp.Reservations))
	for index, reservation := range resp.Reservations {
		instance := reservation.Instances[0]
		ips[index] = *instance.PrivateIpAddress
	}

	sort.Strings(ips)
	for index, ip := range ips {
		if index >= *maxResults {
			break
		}
		fmt.Println(ip)
	}
}
