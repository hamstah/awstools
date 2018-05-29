package main

import (
	"context"
	"net/http"
	"time"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cloudwatch"
)

type Target struct {
	URL            string `json:"url"`
	ExpectedStatus int    `json:"expected_status"`
	Timeout        int    `json:"timeout"`
}

type Cloudwatch struct {
	MetricName string            `json:"metric_name"`
	Dimensions map[string]string `json:"dimensions"`
	Namespace  string            `json:"namespace"`
}

type PingEvent struct {
	Target     Target     `json:"target"`
	Cloudwatch Cloudwatch `json:"cloudwatch"`
}

func DoPing(target Target) (bool, error) {
	timeout := 3
	if target.Timeout > 0 && target.Timeout <= 10 {
		timeout = target.Timeout
	}
	var netClient = &http.Client{
		Timeout: time.Second * time.Duration(timeout),
	}
	response, err := netClient.Get(target.URL)
	if err != nil {
		return false, err
	}
	defer response.Body.Close()
	return response.StatusCode == target.ExpectedStatus, nil
}

func HandleRequest(ctx context.Context, event PingEvent) error {
	svc := cloudwatch.New(session.New())
	cloudwatchDimensions := []*cloudwatch.Dimension{}
	for name, value := range event.Cloudwatch.Dimensions {
		cloudwatchDimensions = append(cloudwatchDimensions, &cloudwatch.Dimension{
			Name:  aws.String(name),
			Value: aws.String(value),
		})
	}

	healthy, err := DoPing(event.Target)
	value := 0.0
	if healthy {
		value = 1.0
	}
	_, err = svc.PutMetricData(&cloudwatch.PutMetricDataInput{
		MetricData: []*cloudwatch.MetricDatum{
			{
				Dimensions: cloudwatchDimensions,
				MetricName: aws.String(event.Cloudwatch.MetricName),
				Value:      aws.Float64(value),
			},
		},
		Namespace: aws.String(event.Cloudwatch.Namespace),
	})
	return err
}

func main() {
	lambda.Start(HandleRequest)
}
