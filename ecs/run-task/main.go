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
	taskDefinition   = kingpin.Flag("task-definition", "ECS task definition").Required().String()
	cluster          = kingpin.Flag("cluster", "ECS cluster").Required().String()
	taskOverrideJSON = kingpin.Flag("task-override-json", "Path to a JSON file with the task override to use").String()
)

func main() {
	kingpin.CommandLine.Name = "ecs-run-task"
	kingpin.CommandLine.Help = "Run a task on ECS."
	flags := common.HandleFlags()

	session, conf := common.OpenSession(flags)

	ecsClient := ecs.New(session, conf)

	taskOverrides, overridesErr := resolveTaskOverride(taskOverrideJSON)
	common.FatalOnError(overridesErr)

	_, err := ecsClient.RunTask(&ecs.RunTaskInput{
		TaskDefinition: taskDefinition,
		Cluster:        cluster,
		Count:          aws.Int64(1),
		Overrides:      taskOverrides,
	})

	common.FatalOnError(err)
}

func resolveTaskOverride(taskOverridesJSON *string) (*ecs.TaskOverride, error) {

	if *taskOverridesJSON == "" {
		return nil, nil
	}

	b, err := os.ReadFile(*taskOverridesJSON)
	if err != nil {
		return nil, err
	}

	var taskOverride *ecs.TaskOverride
	taskOverride = &ecs.TaskOverride{}
	err = json.Unmarshal(b, taskOverride)
	if err != nil {
		return nil, err
	}

	return taskOverride, nil
}
