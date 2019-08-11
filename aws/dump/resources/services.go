package resources

import (
	"fmt"
	"sort"
)

func AllServices() map[string]Service {
	return map[string]Service{
		"acm":         ACMService,
		"autoscaling": AutoScalingService,
		"cloudwatch":  CloudwatchService,
		"ec2":         EC2Service,
		"iam":         IAMService,
		"kms":         KMSService,
		"lambda":      LambdaService,
		"route53":     Route53Service,
		"s3":          S3Service,
		"rds":         RDSService,
	}
}

func AllReports() []string {
	reports := []string{}
	for _, service := range AllServices() {
		for reportName, _ := range service.Reports {
			reports = append(reports, fmt.Sprintf("%s:%s", service.Name, reportName))
		}
	}
	sort.Strings(reports)
	return reports
}
