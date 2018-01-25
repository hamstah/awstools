package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/elb"
	"github.com/aws/aws-sdk-go/service/route53"
)

func main() {

	loadBalancerName := flag.String("name", "", "Name of the load balancer")
	region := flag.String("region", "eu-west-1", "AWS region")
	flag.Parse()

	if len(*loadBalancerName) < 1 {
		fmt.Println("Missing load balancer name")
		os.Exit(1)
	}

	config := aws.Config{Region: aws.String(*region)}
	session := session.New(&config)

	elb_svc := elb.New(session)
	params := &elb.DescribeLoadBalancersInput{
		LoadBalancerNames: []*string{aws.String(*loadBalancerName)},
	}
	resp, err := elb_svc.DescribeLoadBalancers(params)

	if err != nil {
		fmt.Println(err.Error())
		os.Exit(3)
	}

	l := len(resp.LoadBalancerDescriptions)
	if l < 0 {
		fmt.Println("No load balancer found")
		os.Exit(2)
	}

	protocol := "HTTP"
	var port int64 = 80

	for _, listener := range resp.LoadBalancerDescriptions[0].ListenerDescriptions {
		if *listener.Listener.InstanceProtocol == "HTTPS" {
			protocol = *listener.Listener.InstanceProtocol
			port = *listener.Listener.LoadBalancerPort
			break
		}
		port = *listener.Listener.LoadBalancerPort
	}

	dnsName := *resp.LoadBalancerDescriptions[0].DNSName
	dnsNameDot := fmt.Sprintf("%s.", dnsName)
	route53_svc := route53.New(session)

	zones, err := route53_svc.ListHostedZones(&route53.ListHostedZonesInput{})

	if err != nil {
		fmt.Println(err.Error())
		os.Exit(3)
	}
	for _, hostedZone := range zones.HostedZones {
		zoneId := hostedZone.Id

		records, err := route53_svc.ListResourceRecordSets(&route53.ListResourceRecordSetsInput{HostedZoneId: zoneId})
		if err != nil {
			fmt.Println(err.Error())
			break
		}
		for _, record := range records.ResourceRecordSets {
			if record.AliasTarget == nil || record.AliasTarget.DNSName == nil {
				continue
			}
			if *record.AliasTarget.DNSName == dnsNameDot {
				dnsName = strings.TrimRight(*record.Name, ".")
				break
			}

		}
	}
	fmt.Println(fmt.Sprintf("%s://%s:%d", strings.ToLower(protocol), dnsName, port))
}
