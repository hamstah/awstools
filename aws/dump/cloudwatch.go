package main

import "github.com/aws/aws-sdk-go/service/cloudwatch"

var (
	CloudwatchService = Service{
		Name: "cloudwatch",
		Reports: map[string]Report{
			"alarms": CloudwatchListAlarms,
		},
	}
)

func CloudwatchListAlarms(session *Session) *ReportResult {
	client := cloudwatch.New(session.Session, session.Config)

	result := &ReportResult{}
	result.Error = client.DescribeAlarmsPages(&cloudwatch.DescribeAlarmsInput{},
		func(page *cloudwatch.DescribeAlarmsOutput, lastPage bool) bool {
			for _, alarm := range page.MetricAlarms {

				resource, err := NewResource(*alarm.AlarmArn, alarm)
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
