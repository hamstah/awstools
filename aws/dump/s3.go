package main

import (
	"fmt"

	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/fatih/structs"
)

var (
	S3Service = Service{
		Name: "s3",
		Reports: map[string]Report{
			"buckets": S3ListBuckets,
		},
	}
)

func S3ListBuckets(session *Session) *ReportResult {
	client := s3.New(session.Session, session.Config)

	buckets := []Resource{}

	res, err := client.ListBuckets(&s3.ListBucketsInput{})
	if err != nil {
		return &ReportResult{nil, err}
	}

	for _, bucket := range res.Buckets {
		buckets = append(buckets, Resource{
			ID:        *bucket.Name,
			ARN:       fmt.Sprintf("arn:aws:s3:::%s", *bucket.Name),
			AccountID: session.AccountID,
			Service:   "s3",
			Type:      "bucket",
			// AccountID
			// Region: Need to use GetBucketLocation
			Metadata: structs.Map(bucket),
		})
	}

	return &ReportResult{buckets, err}
}
