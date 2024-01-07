package main

import (
	"encoding/json"
	"os"

	kingpin "github.com/alecthomas/kingpin/v2"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ecs"
	"github.com/hamstah/awstools/common"
)

var (
	taskDefinition    = kingpin.Flag("task-definition", "ECS task definition").Required().String()
	cluster           = kingpin.Flag("cluster", "ECS cluster").Required().String()
	taskOverridesJSON = kingpin.Flag("task-overrides-json", "Path to a JSON file with the task overrides to use").String()
)

func main() {
	kingpin.CommandLine.Name = "ecs-run-task"
	kingpin.CommandLine.Help = "Run a task on ECS."
	flags := common.HandleFlags()

	session, conf := common.OpenSession(flags)

	ecsClient := ecs.New(session, conf)

	taskOverrides, err := resolveTaskOverrides(*taskOverridesJSON)
	common.FatalOnError(err)

	_, err = ecsClient.RunTask(&ecs.RunTaskInput{
		TaskDefinition: taskDefinition,
		Cluster:        cluster,
		Count:          aws.Int64(1),
		Overrides:      taskOverrides,
	})

	common.FatalOnError(err)
}

func resolveTaskOverrides(taskOverridesJSON string) (*ecs.TaskOverride, error) {
	if taskOverridesJSON == "" {
		return nil, nil
	}

	b, err := os.ReadFile(taskOverridesJSON)
	if err != nil {
		return nil, err
	}

	taskOverrides := &ecs.TaskOverride{}
	if err := json.Unmarshal(b, taskOverrides); err != nil {
		return nil, err
	}

	return taskOverrides, nil
}
