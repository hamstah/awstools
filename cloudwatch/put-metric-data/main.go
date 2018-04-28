package main

import (
	"github.com/aws/aws-sdk-go/service/cloudwatch"
	"github.com/hamstah/awstools/common"
	kingpin "gopkg.in/alecthomas/kingpin.v2"
)

var (
	flags      = common.KingpinSessionFlags()
	metricName = kingpin.Flag("metric-name", "Name of the Cloudwatch metric").Required().String()
	namespace  = kingpin.Flag("namespace", "Name of the Cloudwatch namespace").Required().String()
	value      = kingpin.Flag("value", "Metric value").Required().Float64()
)

func main() {
	kingpin.CommandLine.Name = "cloudwatch-put-metric-data"
	kingpin.CommandLine.Help = "Put a cloudwatch metric value."
	kingpin.Parse()

	session, conf := common.OpenSession(flags)

	cloudwatchClient := cloudwatch.New(session, conf)
	_, err := cloudwatchClient.PutMetricData(&cloudwatch.PutMetricDataInput{
		MetricData: []*cloudwatch.MetricDatum{
			{
				MetricName: metricName,
				Value:      value,
			},
		},
		Namespace: namespace,
	})
	common.FatalOnError(err)
}
