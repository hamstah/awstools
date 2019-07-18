package main

import (
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go/service/elb"
	"github.com/aws/aws-sdk-go/service/route53"
	"github.com/hamstah/awstools/common"
	kingpin "gopkg.in/alecthomas/kingpin.v2"
)

var (
	loadBalancerName = kingpin.Flag("name", "Name of the load balancer").Required().String()
)

func main() {

	kingpin.CommandLine.Name = "elb-resolve-elb-external-url"
	kingpin.CommandLine.Help = "Resolve the public URL of an ELB."
	flags := common.HandleFlags()

	session, conf := common.OpenSession(flags)

	elbClient := elb.New(session, conf)

	resp, err := elbClient.DescribeLoadBalancers(&elb.DescribeLoadBalancersInput{
		LoadBalancerNames: []*string{loadBalancerName},
	})
	common.FatalOnError(err)

	l := len(resp.LoadBalancerDescriptions)
	if l == 0 {
		common.Fatalln("No load balancer found")
	}

	protocol := "HTTP"
	port := int64(80)

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
	fmt.Println(fmt.Sprintf("%s://%s:%d", strings.ToLower(protocol), dnsName, port))
}
