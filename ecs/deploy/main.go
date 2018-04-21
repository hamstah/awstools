package main

import (
	"fmt"
	"os"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ecs"
	"github.com/hamstah/awstools/common"

	kingpin "gopkg.in/alecthomas/kingpin.v2"
)

var (
	flags    = common.KingpinSessionFlags()
	taskName = kingpin.Flag("task-name", "ECS task name").Required().String()
	cluster  = kingpin.Flag("cluster", "ECS cluster").Required().String()
	services = kingpin.Flag("service", "ECS services").Required().Strings()
	images   = kingpin.Flag("image", "Change the images to the new ones. Container name=image").StringMap()
	timeout  = kingpin.Flag("timeout", "Timeout when waiting for services to update").Default("300s").Duration()
)

func main() {
	kingpin.CommandLine.Name = "ecs-deploy"
	kingpin.CommandLine.Help = "Update a task definition on ECS."
	kingpin.Parse()

	session := session.Must(session.NewSession())
	conf := common.AssumeRoleConfig(flags, session)

	ecsClient := ecs.New(session, conf)

	taskDefinition, err := getTaskDefinition(ecsClient, taskName)
	common.FatalOnError(err)

	if len(*images) != 0 {
		for _, containerDefinition := range taskDefinition.ContainerDefinitions {
			newImage := (*images)[*containerDefinition.Name]
			if newImage != "" {
				containerDefinition.Image = &newImage
			}
		}
	}

	newTaskDefinition, err := updateTaskDefinition(ecsClient, taskDefinition)
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
			pending += 1
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
	previousPending := 0
	for pending > 0 {
		servicesResult, err := ecsClient.DescribeServices(servicesInput)
		common.FatalOnError(err)

		previousPending = pending
		pending = 0
		for _, service := range servicesResult.Services {
			if *service.Deployments[0].PendingCount != 0 {
				pending += 1
			}
		}

		if pending != 0 {
			if time.Since(start) >= *timeout {
				common.Fatalln(fmt.Sprintf("%d still pending, giving up after %s", pending, *timeout))
			}
			if previousPending != pending {
				fmt.Println(fmt.Sprintf("Waiting for %d service(s) to become ready", pending))
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

func updateTaskDefinition(ecsClient *ecs.ECS, taskDefinition *ecs.TaskDefinition) (*ecs.TaskDefinition, error) {
	updateInput := &ecs.RegisterTaskDefinitionInput{
		ContainerDefinitions:    taskDefinition.ContainerDefinitions,
		Cpu:                     taskDefinition.Cpu,
		ExecutionRoleArn:        taskDefinition.ExecutionRoleArn,
		Family:                  taskDefinition.Family,
		Memory:                  taskDefinition.Memory,
		NetworkMode:             taskDefinition.NetworkMode,
		PlacementConstraints:    taskDefinition.PlacementConstraints,
		RequiresCompatibilities: taskDefinition.RequiresCompatibilities,
		TaskRoleArn:             taskDefinition.TaskRoleArn,
		Volumes:                 taskDefinition.Volumes,
	}

	updateResult, err := ecsClient.RegisterTaskDefinition(updateInput)
	if err != nil {
		return nil, err
	}

	return updateResult.TaskDefinition, nil
}
