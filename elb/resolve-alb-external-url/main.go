package main

import (
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/elbv2"
	"github.com/aws/aws-sdk-go/service/route53"
	"github.com/hamstah/awstools/common"
	kingpin "gopkg.in/alecthomas/kingpin.v2"
)

var (
	flags            = common.KingpinSessionFlags()
	loadBalancerName = kingpin.Flag("name", "Name of the load balancer").Required().String()
)

func main() {
	kingpin.CommandLine.Name = "elb-resolve-alb-external-url"
	kingpin.CommandLine.Help = "Resolve the public URL of an ALB."
	kingpin.Parse()

	session := session.Must(session.NewSession())
	conf := common.AssumeRoleConfig(flags, session)

	elbClient := elbv2.New(session, conf)
	resp, err := elbClient.DescribeLoadBalancers(&elbv2.DescribeLoadBalancersInput{
		Names: []*string{loadBalancerName},
	})
	common.FatalOnError(err)

	l := len(resp.LoadBalancers)
	if l == 0 {
		common.Fatalln("No load balancer found")
	}

	dnsName := *resp.LoadBalancers[0].DNSName
	dnsNameDot := fmt.Sprintf("%s.", dnsName)

	route53Client := route53.New(session, conf)
	zones, err := route53Client.ListHostedZones(&route53.ListHostedZonesInput{})
	common.FatalOnError(err)

	for _, hostedZone := range zones.HostedZones {
		records, err := route53Client.ListResourceRecordSets(&route53.ListResourceRecordSetsInput{
			HostedZoneId: hostedZone.Id,
		})
		common.FatalOnError(err)

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
	fmt.Println(dnsName)
}
