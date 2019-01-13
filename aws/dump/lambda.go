package main

import (
	"github.com/aws/aws-sdk-go/service/lambda"
)

var (
	LambdaService = Service{
		Name: "lambda",
		Reports: map[string]Report{
			"functions": LambdaListFunctions,
		},
	}
)

func LambdaListFunctions(session *Session) *ReportResult {
	client := lambda.New(session.Session, session.Config)

	result := &ReportResult{}
	result.Error = client.ListFunctionsPages(&lambda.ListFunctionsInput{},
		func(page *lambda.ListFunctionsOutput, lastPage bool) bool {
			for _, function := range page.Functions {
				resource, err := NewResource(*function.FunctionArn, function)
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
