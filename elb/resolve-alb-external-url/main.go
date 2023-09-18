package main

import (
	"fmt"
	"os"
	"strings"

	kingpin "github.com/alecthomas/kingpin/v2"
	"github.com/aws/aws-sdk-go/service/elbv2"
	"github.com/aws/aws-sdk-go/service/route53"
	"github.com/hamstah/awstools/common"
)

var (
	loadBalancerName = kingpin.Flag("name", "Name of the load balancer").Required().String()
	dnsPrefix        = kingpin.Flag("dns-prefix", "Prefix to match on the DNS").String()
)

func main() {
	kingpin.CommandLine.Name = "elb-resolve-alb-external-url"
	kingpin.CommandLine.Help = "Resolve the public URL of an ALB."
	flags := common.HandleFlags()

	session, conf := common.OpenSession(flags)

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
	found := false
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
				trimmedDNSName := strings.TrimRight(*record.Name, ".")
				if *dnsPrefix != "" && strings.HasPrefix(trimmedDNSName, *dnsPrefix) {
					dnsName = trimmedDNSName
					found = true
					break
				}
			}
		}
		if found {
			break
		}
	}
	if !found {
		os.Exit(1)
	}

	fmt.Println(dnsName)
}
