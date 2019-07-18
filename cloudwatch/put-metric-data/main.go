package main

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudwatch"
	"github.com/hamstah/awstools/common"
	kingpin "gopkg.in/alecthomas/kingpin.v2"
)

var (
	metricName = kingpin.Flag("metric-name", "Name of the Cloudwatch metric").Required().String()
	namespace  = kingpin.Flag("namespace", "Name of the Cloudwatch namespace").Required().String()
	dimensions = kingpin.Flag("dimension", "Dimensions name=value").StringMap()
	value      = kingpin.Flag("value", "Metric value").Required().Float64()
)

func main() {
	kingpin.CommandLine.Name = "cloudwatch-put-metric-data"
	kingpin.CommandLine.Help = "Put a cloudwatch metric value."
	flags := common.HandleFlags()

	session, conf := common.OpenSession(flags)

	cloudwatchDimensions := []*cloudwatch.Dimension{}
	for name, value := range *dimensions {
		cloudwatchDimensions = append(cloudwatchDimensions, &cloudwatch.Dimension{
			Name:  aws.String(name),
			Value: aws.String(value),
		})
	}

	cloudwatchClient := cloudwatch.New(session, conf)
	_, err := cloudwatchClient.PutMetricData(&cloudwatch.PutMetricDataInput{
		MetricData: []*cloudwatch.MetricDatum{
			{
				Dimensions: cloudwatchDimensions,
				MetricName: metricName,
				Value:      value,
			},
		},
		Namespace: namespace,
	})
	common.FatalOnError(err)
}
