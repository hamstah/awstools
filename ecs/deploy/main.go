package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"time"

	kingpin "github.com/alecthomas/kingpin/v2"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ecs"
	"github.com/aws/aws-sdk-go/service/sts"
	"github.com/hamstah/awstools/common"
)

var (
	taskName            = kingpin.Flag("task-name", "ECS task name").String()
	cluster             = kingpin.Flag("cluster", "ECS cluster").Required().String()
	services            = kingpin.Flag("service", "ECS services").Required().Strings()
	images              = kingpin.Flag("image", "Change the images to the new ones. Format is container_name=image. Can be repeated.").StringMap()
	timeout             = kingpin.Flag("timeout", "Timeout when waiting for services to update").Default("300s").Duration()
	taskJSON            = kingpin.Flag("task-json", "Path to a JSON file with the task definition to use").String()
	taskVariables       = kingpin.Flag("task-variables", "Variables to be replaced in the task definition").StringMap()
	overwriteAccountIDs = kingpin.Flag("overwrite-account-ids", "Overwrite account IDs in role ARN with the caller account ID").Default("false").Bool()
)

func main() {
	kingpin.CommandLine.Name = "ecs-deploy"
	kingpin.CommandLine.Help = "Update a task definition on ECS."
	flags := common.HandleFlags()

	session, conf := common.OpenSession(flags)

	ecsClient := ecs.New(session, conf)

	if *taskJSON != "" && *taskName != "" {
		common.Fatalln("Use only one of --task-json and --task-name")
	}

	var err error
	var taskDefinition *ecs.TaskDefinition

	switch {
	case *taskJSON != "":
		b, err := os.ReadFile(*taskJSON)
		common.FatalOnError(err)

		if taskVariables != nil {
			for k, v := range *taskVariables {
				b = bytes.ReplaceAll(b, []byte(fmt.Sprintf("${%s}", k)), []byte(v))
			}
		}

		taskDefinition = &ecs.TaskDefinition{}
		err = json.Unmarshal(b, taskDefinition)
		common.FatalOnError(err)

		taskName = taskDefinition.Family
	case *taskName != "":
		taskDefinition, err = getTaskDefinition(ecsClient, taskName)
		common.FatalOnError(err)
	default:
		common.Fatalln("One of --task-json or --task-name is required")
	}

	if len(*images) != 0 {
		for _, containerDefinition := range taskDefinition.ContainerDefinitions {
			newImage := (*images)[*containerDefinition.Name]
			if newImage != "" {
				containerDefinition.Image = &newImage
			}
		}
	}

	accountID := ""
	if *overwriteAccountIDs {
		stsClient := sts.New(session, conf)
		res, err := stsClient.GetCallerIdentity(&sts.GetCallerIdentityInput{})
		common.FatalOnError(err)
		accountID = *res.Account
	}

	newTaskDefinition, err := updateTaskDefinition(ecsClient, taskDefinition, accountID)
	common.FatalOnError(err)

	fmt.Println(*newTaskDefinition.TaskDefinitionArn)

	pending := 0
	for _, service := range *services {
		_, err := ecsClient.UpdateService(&ecs.UpdateServiceInput{
			Cluster:        cluster,
			Service:        aws.String(service),
			TaskDefinition: newTaskDefinition.TaskDefinitionArn,
		})
		if err != nil {
			fmt.Fprintln(os.Stderr, "Failed to update service", service, err)
		} else {
			pending++
		}
	}

	serviceNamesInput := []*string{}
	for _, service := range *services {
		serviceNamesInput = append(serviceNamesInput, aws.String(service))
	}

	servicesInput := &ecs.DescribeServicesInput{
		Cluster:  cluster,
		Services: serviceNamesInput,
	}

	start := time.Now()
	for pending > 0 {
		servicesResult, err := ecsClient.DescribeServices(servicesInput)
		common.FatalOnError(err)

		previousPending := pending
		pending = 0
		for _, service := range servicesResult.Services {
			if *service.Deployments[0].RunningCount != *service.Deployments[0].DesiredCount {
				pending++
			}
		}

		if pending != 0 {
			if time.Since(start) >= *timeout {
				common.Fatalln(fmt.Sprintf("%d still pending, giving up after %s", pending, *timeout))
			}
			if previousPending != pending {
				fmt.Printf("Waiting for %d service(s) to become ready\n", pending)
			}
			time.Sleep(1 * time.Second)
		}
	}
}

func getTaskDefinition(ecsClient *ecs.ECS, taskName *string) (*ecs.TaskDefinition, error) {
	input := &ecs.DescribeTaskDefinitionInput{
		TaskDefinition: taskName,
	}

	result, err := ecsClient.DescribeTaskDefinition(input)
	if err != nil {
		return nil, err
	}
	return result.TaskDefinition, nil
}

func updateTaskDefinition(ecsClient *ecs.ECS, taskDefinition *ecs.TaskDefinition, accountID string) (*ecs.TaskDefinition, error) {
	var err error
	taskRoleArn := taskDefinition.TaskRoleArn
	executionRoleArn := taskDefinition.ExecutionRoleArn

	if accountID != "" {
		taskRoleArn, err = common.ReplaceAccountIDPtr(taskRoleArn, accountID)
		if err != nil {
			return nil, fmt.Errorf("failed to replace accountID in task role arn")
		}

		executionRoleArn, err = common.ReplaceAccountIDPtr(executionRoleArn, accountID)
		if err != nil {
			return nil, fmt.Errorf("failed to replace accountID in execution role arn")
		}
	}

	updateInput := &ecs.RegisterTaskDefinitionInput{
		ContainerDefinitions:    taskDefinition.ContainerDefinitions,
		Cpu:                     taskDefinition.Cpu,
		ExecutionRoleArn:        executionRoleArn,
		Family:                  taskDefinition.Family,
		Memory:                  taskDefinition.Memory,
		NetworkMode:             taskDefinition.NetworkMode,
		PlacementConstraints:    taskDefinition.PlacementConstraints,
		RequiresCompatibilities: taskDefinition.RequiresCompatibilities,
		TaskRoleArn:             taskRoleArn,
		Volumes:                 taskDefinition.Volumes,
	}

	updateResult, err := ecsClient.RegisterTaskDefinition(updateInput)
	if err != nil {
		return nil, err
	}

	return updateResult.TaskDefinition, nil
}
