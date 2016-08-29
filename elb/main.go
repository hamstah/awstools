package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/elb"
)

func main() {

	loadBalancerName := flag.String("name", "", "Name of the load balancer")
	region := flag.String("region", "eu-west-1", "AWS region")
	flag.Parse()

	if len(*loadBalancerName) < 1 {
		fmt.Println("Missing load balancer name")
		os.Exit(1)
	}

	config := aws.Config{Region: aws.String("eu-west-1")}

	svc := elb.New(session.New(&config))

	params := &elb.DescribeLoadBalancersInput{
		LoadBalancerNames: []*string{aws.String(*loadBalancerName)},
	}
	resp, err := svc.DescribeLoadBalancers(params)

	if err != nil {
		fmt.Println(err.Error())
		os.Exit(3)
	}

	l := len(resp.LoadBalancerDescriptions)
	if l < 0 {
		fmt.Println("No load balancer found")
		os.Exit(2)
	}

	fmt.Println(*resp.LoadBalancerDescriptions[0].DNSName)
}
