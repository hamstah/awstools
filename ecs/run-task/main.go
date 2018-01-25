package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ecs"
)

func main() {

	taskDefinitionName := flag.String("task-definition", "", "Task definition name")
	clusterName := flag.String("cluster-name", "", "Cluster name")
	region := flag.String("region", "eu-west-1", "AWS region")
	flag.Parse()

	if len(*taskDefinitionName) < 1 {
		fmt.Println("Missing task definition name")
		os.Exit(1)
	}

	if len(*clusterName) < 1 {
		fmt.Println("Missing cluster name")
		os.Exit(1)
	}


	config := aws.Config{Region: aws.String(*region)}
	session := session.New(&config)

	svc := ecs.New(session)

	params := &ecs.RunTaskInput{
		TaskDefinition: aws.String(*taskDefinitionName),
		Cluster:        aws.String(*clusterName),
		Count:          aws.Int64(1),
	}
	_, err := svc.RunTask(params)

	if err != nil {
		fmt.Println(err.Error())
		os.Exit(2)
	}
}
