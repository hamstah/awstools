package main

import (
	"fmt"
	"os"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ecs"

	kingpin "gopkg.in/alecthomas/kingpin.v2"
)

var (
	app         = kingpin.New("ecs-deploy", "Update a task definition on ECS")
	region      = app.Flag("region", "AWS Region").Default("eu-west-1").String()
	taskName    = app.Flag("task-name", "ECS task name").Required().String()
	clusterName = app.Flag("cluster", "ECS cluster").Required().String()
	services    = app.Flag("service", "ECS services").Required().Strings()
	images      = app.Flag("images", "Change the images to the new ones. Container name=image").StringMap()
	timeout     = app.Flag("timeout", "Timeout when waiting for services to update").Default("300s").Duration()
)


func getTaskDefinition(svc *ecs.ECS, taskName string) (*ecs.TaskDefinition) {
	input := &ecs.DescribeTaskDefinitionInput{
		TaskDefinition: aws.String(taskName),
	}

	result, err := svc.DescribeTaskDefinition(input)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Failed to fetch the task definition", err)
		os.Exit(1)
	}
	return result.TaskDefinition
}

func updateTaskDefinition(svc *ecs.ECS, taskDefinition *ecs.TaskDefinition) (*ecs.TaskDefinition) {
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

	updateResult, err := svc.RegisterTaskDefinition(updateInput)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Failed to update the task definition", err)
		os.Exit(1)
	}
	return updateResult.TaskDefinition
}

func main() {
	kingpin.MustParse(app.Parse(os.Args[1:]))

	config := aws.Config{Region: aws.String(*region)}
	session := session.New(&config)
	svc := ecs.New(session)

	taskDefinition := getTaskDefinition(svc, *taskName)

	if len(*images) != 0 {
		for _, containerDefinition := range taskDefinition.ContainerDefinitions {
			newImage := (*images)[*containerDefinition.Name]
			if newImage != "" {
				containerDefinition.Image = &newImage
			}
		}
	}

	newTaskDefinition := updateTaskDefinition(svc, taskDefinition)
	fmt.Println(*newTaskDefinition.TaskDefinitionArn)

	pending := 0
	for _, service := range *services {

		updateServiceInput := &ecs.UpdateServiceInput{
			Cluster:        aws.String(*clusterName),
			Service:        aws.String(service),
			TaskDefinition: aws.String(*newTaskDefinition.TaskDefinitionArn),
		}
		_, err := svc.UpdateService(updateServiceInput)
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
		Cluster:  aws.String(*clusterName),
		Services: serviceNamesInput,
	}

	start := time.Now()
	previousPending := 0
	for pending > 0 {
		servicesResult, err := svc.DescribeServices(servicesInput)
		if err != nil {
			fmt.Fprintln(os.Stderr, "Failed to fetch services", err)
			os.Exit(2)
		}

		previousPending = pending
		pending = 0
		for _, service := range servicesResult.Services {

			if *service.Deployments[0].PendingCount != 0 {
				pending += 1
			}
		}

		if pending != 0 {
			if time.Since(start) >= *timeout {
				fmt.Println(os.Stderr, fmt.Sprintf("%s still pending, giving up after %s", pending, *timeout))
				os.Exit(3)
			}
			if previousPending != pending {
				fmt.Println(fmt.Sprintf("Waiting for %d service(s) to become ready", pending))
			}
			time.Sleep(1 * time.Second)
		}
	}
}
