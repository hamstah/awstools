package resources

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
	}	
}