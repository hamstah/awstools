package main

import (
	kingpin "github.com/alecthomas/kingpin/v2"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ecs"
	"github.com/hamstah/awstools/common"
)

var (
	taskDefinition = kingpin.Flag("task-definition", "ECS task definition").Required().String()
	cluster        = kingpin.Flag("cluster", "ECS cluster").Required().String()
)

func main() {
	kingpin.CommandLine.Name = "ecs-run-task"
	kingpin.CommandLine.Help = "Run a task on ECS."
	flags := common.HandleFlags()

	session, conf := common.OpenSession(flags)

	ecsClient := ecs.New(session, conf)

	_, err := ecsClient.RunTask(&ecs.RunTaskInput{
		TaskDefinition: taskDefinition,
		Cluster:        cluster,
		Count:          aws.Int64(1),
	})
	common.FatalOnError(err)
}
