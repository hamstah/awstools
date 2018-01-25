package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cloudwatch"
)

func main() {

	metricName := flag.String("metric-name", "", "Metric name")
	namespace := flag.String("namespace", "", "Namespace")
	value := flag.Float64("value", 0.0, "Value")
	region := flag.String("region", "eu-west-1", "AWS region")
	flag.Parse()

	if len(*metricName) < 1 {
		fmt.Println("Missing task metric name")
		os.Exit(1)
	}

	if len(*namespace) < 1 {
		fmt.Println("Missing namespace")
		os.Exit(1)
	}


	config := aws.Config{Region: aws.String(*region)}
	session := session.New(&config)

	svc := cloudwatch.New(session)

	params := &cloudwatch.PutMetricDataInput{
		MetricData: []*cloudwatch.MetricDatum{
			{ // Required
				MetricName: aws.String(*metricName), // Required
				// Dimensions: []*cloudwatch.Dimension{
				// 	{ // Required
				// 		Name:  aws.String("DimensionName"),  // Required
				// 		Value: aws.String("DimensionValue"), // Required
				// 	},
				// 	// More values...
				// },
				// StatisticValues: &cloudwatch.StatisticSet{
				// 	Maximum:     aws.Float64(1.0), // Required
				// 	Minimum:     aws.Float64(1.0), // Required
				// 	SampleCount: aws.Float64(1.0), // Required
				// 	Sum:         aws.Float64(1.0), // Required
				// },
				// Timestamp: aws.Time(time.Now()),
				// Unit:      aws.String("StandardUnit"),
				Value: aws.Float64(*value),
			},
			// More values...
		},
		Namespace: aws.String(*namespace),
	}
	_, err := svc.PutMetricData(params)

	if err != nil {
		// Print the error, cast err to awserr.Error to get the Code and
		// Message from an error.
		fmt.Println(err.Error())
		os.Exit(2)
	}
}
