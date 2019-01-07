package main

import "github.com/aws/aws-sdk-go/service/cloudwatch"

func CloudwatchListAlarms(session *Session) *FetchResult {
	client := cloudwatch.New(session.Session, session.Config)

	result := &FetchResult{}
	result.Error = client.DescribeAlarmsPages(&cloudwatch.DescribeAlarmsInput{},
		func(page *cloudwatch.DescribeAlarmsOutput, lastPage bool) bool {
			for _, alarm := range page.MetricAlarms {

				resource, err := NewResource(*alarm.AlarmArn)
				if err != nil {
					result.Error = err
					return false
				}
				result.Resources = append(result.Resources, *resource)
			}

			return true
		})

	return result
}
