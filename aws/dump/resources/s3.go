package resources

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

	result := &ReportResult{[]Resource{}, nil}
	res, err := client.ListBuckets(&s3.ListBucketsInput{})
	if err != nil {
		return &ReportResult{nil, err}
	}

	for _, bucket := range res.Buckets {

		location, err := client.GetBucketLocation(&s3.GetBucketLocationInput{
			Bucket: bucket.Name,
		})

		if err != nil {
			result.Error = err
			return result
		}

		if *location.LocationConstraint != *session.Config.Region {
			continue
		}

		result.Resources = append(result.Resources, Resource{
			ID:        *bucket.Name,
			ARN:       fmt.Sprintf("arn:aws:s3:::%s", *bucket.Name),
			AccountID: session.AccountID,
			Service:   "s3",
			Type:      "bucket",
			Region:    *location.LocationConstraint,
			Metadata:  structs.Map(bucket),
		})

		policy, err := client.GetBucketPolicy(&s3.GetBucketPolicyInput{
			Bucket: bucket.Name,
		})
		if err != nil {
			result.Error = err
			return result
		}

		document, err := DecodeInlinePolicyDocument(*policy.Policy)
		if err != nil {
			result.Error = err
			return result
		}

		result.Resources = append(result.Resources, Resource{
			ID:        *bucket.Name,
			AccountID: session.AccountID,
			Service:   "s3",
			Type:      "bucket-policy",
			Metadata: map[string]interface{}{
				"PolicyDocument": document,
			},
		})

	}

	return result
}
