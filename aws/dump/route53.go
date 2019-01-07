package main

import "github.com/aws/aws-sdk-go/service/route53"

func Route53ListHostedZones(session *Session) *FetchResult {
	client := route53.New(session.Session, session.Config)

	result := &FetchResult{}
	result.Error = client.ListHostedZonesPages(&route53.ListHostedZonesInput{},
		func(page *route53.ListHostedZonesOutput, lastPage bool) bool {
			for _, zone := range page.HostedZones {
				resource := &Resource{
					ID:        *zone.Id,
					AccountID: session.AccountID,
					Service:   "route53",
					Type:      "hosted-zone",
					Metadata: map[string]string{
						"Name": *zone.Name,
					},
				}
				result.Resources = append(result.Resources, *resource)
			}

			return true
		})

	return result
}
