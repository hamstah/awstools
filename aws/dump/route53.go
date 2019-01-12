package main

import (
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/route53"
)

func Route53ListHostedZonesAndRecordSets(session *Session) *FetchResult {
	client := route53.New(session.Session, session.Config)

	result := &FetchResult{}
	result.Error = client.ListHostedZonesPages(&route53.ListHostedZonesInput{},
		func(page *route53.ListHostedZonesOutput, lastPage bool) bool {
			for _, zone := range page.HostedZones {

				parts := strings.Split(*zone.Id, "/")
				shortID := parts[len(parts)-1]

				resource := &Resource{
					ID:        shortID,
					AccountID: session.AccountID,
					Service:   "route53",
					Type:      "zone",
					Metadata: map[string]interface{}{
						"Name": *zone.Name,
					},
				}
				result.Resources = append(result.Resources, *resource)

				records := Route53ListResourceRecordSets(session, *zone.Id)
				if records.Error != nil {
					result.Error = records.Error
					return false
				}
				result.Resources = append(result.Resources, records.Resources...)
			}

			return true
		})

	return result
}

func Route53ListResourceRecordSets(session *Session, hostedZoneID string) *FetchResult {
	client := route53.New(session.Session, session.Config)

	parts := strings.Split(hostedZoneID, "/")
	shortID := parts[len(parts)-1]

	result := &FetchResult{}
	result.Error = client.ListResourceRecordSetsPages(&route53.ListResourceRecordSetsInput{HostedZoneId: aws.String(hostedZoneID)},
		func(page *route53.ListResourceRecordSetsOutput, lastPage bool) bool {
			for _, set := range page.ResourceRecordSets {
				if *set.Type == "NS" || *set.Type == "SOA" {
					continue
				}

				records := []string{}
				for _, record := range set.ResourceRecords {
					records = append(records, *record.Value)
				}

				resource := &Resource{
					ID:        fmt.Sprintf("%s_%s_%s", shortID, strings.TrimRight(*set.Name, "."), *set.Type),
					AccountID: session.AccountID,
					Service:   "route53",
					Type:      "record",
					Metadata: map[string]interface{}{
						"Name": *set.Name,
						"Type": *set.Type,
						"HostedZoneId": shortID,
					},
				}

				if len(records) > 0 {
					resource.Metadata["ResourceRecords"] = records
				}

				if set.AliasTarget != nil {
					resource.Metadata["AliasTarget"] = *set.AliasTarget
				}

				if set.TTL != nil {
					resource.Metadata["Ttl"] = fmt.Sprintf("%d", *set.TTL)
				}
				result.Resources = append(result.Resources, *resource)
			}

			return true
		})


	return result
}
