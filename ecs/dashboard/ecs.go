package main

import (
	"github.com/aws/aws-sdk-go/service/ecs"
)

func getTaskDefinition(svc *ecs.ECS, taskName *string) (*ecs.TaskDefinition, error) {
	input := &ecs.DescribeTaskDefinitionInput{
		TaskDefinition: taskName,
	}

	result, err := svc.DescribeTaskDefinition(input)
	if err != nil {
		return nil, err
	}
	return result.TaskDefinition, nil
}

func getClusters(svc *ecs.ECS) ([]*string, error) {
	input := &ecs.ListClustersInput{}

	result, err := svc.ListClusters(input)
	if err != nil {
		return nil, err
	}
	return result.ClusterArns, nil
}

func describeClusters(svc *ecs.ECS, clusterArns []*string) ([]*ecs.Cluster, error) {
	input := &ecs.DescribeClustersInput{
		Clusters: clusterArns,
	}

	result, err := svc.DescribeClusters(input)
	if err != nil {
		return nil, err
	}

	return result.Clusters, nil
}

func listServices(svc *ecs.ECS, clusterArn *string) ([]*string, error) {
	input := &ecs.ListServicesInput{
		Cluster: clusterArn,
	}

	result := []*string{}

	err := svc.ListServicesPages(input,
		func(page *ecs.ListServicesOutput, lastPage bool) bool {
			result = append(result, page.ServiceArns...)
			return true
		})
	if err != nil {
		return nil, err
	}
	return result, nil
}

func describeServices(svc *ecs.ECS, clusterName *string, serviceNames []*string) ([]*ecs.Service, error) {
	input := &ecs.DescribeServicesInput{
		Cluster:  clusterName,
		Services: serviceNames,
	}
	result, err := svc.DescribeServices(input)
	if err != nil {
		return nil, err
	}
	return result.Services, nil
}
