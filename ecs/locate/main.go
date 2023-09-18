package main

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/ecs"
	"github.com/hamstah/awstools/common"

	kingpin "github.com/alecthomas/kingpin/v2"
)

var (
	containerName = kingpin.Flag("container-name", "ECS container name").Required().String()
	containerPort = kingpin.Flag("container-port", "ECS container port").Required().Int64()
	cluster       = kingpin.Flag("cluster", "ECS cluster").Required().String()
	service       = kingpin.Flag("service", "ECS service").Required().String()
)

func main() {
	kingpin.CommandLine.Name = "ecs-locate"
	kingpin.CommandLine.Help = "Find an instance/port for a service"
	flags := common.HandleFlags()

	session, conf := common.OpenSession(flags)

	ecsClient := ecs.New(session, conf)

	tasksResult, err := ecsClient.ListTasks(&ecs.ListTasksInput{
		Cluster:       cluster,
		DesiredStatus: aws.String("RUNNING"),
		ServiceName:   service,
	})
	common.FatalOnError(err)

	describeTasks, err := ecsClient.DescribeTasks(&ecs.DescribeTasksInput{
		Cluster: cluster,
		Tasks:   tasksResult.TaskArns,
	})
	common.FatalOnError(err)

	containerInstances := map[string]bool{}

	bindings := map[string]int64{}
	for _, task := range describeTasks.Tasks {
		found := false
		for _, container := range task.Containers {
			if *container.Name != *containerName {
				continue
			}

			for _, binding := range container.NetworkBindings {
				if *binding.ContainerPort != *containerPort {
					continue
				}

				bindings[*task.ContainerInstanceArn] = *binding.HostPort
				containerInstances[*task.ContainerInstanceArn] = true
				found = true
			}
		}
		if !found {
			common.Fatalln(fmt.Sprintf("Could not find container in task %s", *task.TaskArn))
		}
	}

	containerInstancesArns := make([]*string, 0, len(containerInstances))
	for key := range containerInstances {
		containerInstancesArns = append(containerInstancesArns, aws.String(key))
	}

	containerInstancesResult, err := ecsClient.DescribeContainerInstances(&ecs.DescribeContainerInstancesInput{
		Cluster:            cluster,
		ContainerInstances: containerInstancesArns,
	})
	common.FatalOnError(err)

	containerInstanceToEC2 := map[string]string{}
	ec2InstanceIDs := make([]*string, 0, len(containerInstancesResult.ContainerInstances))
	for _, containerInstance := range containerInstancesResult.ContainerInstances {
		ec2InstanceIDs = append(ec2InstanceIDs, containerInstance.Ec2InstanceId)
		containerInstanceToEC2[*containerInstance.ContainerInstanceArn] = *containerInstance.Ec2InstanceId
	}

	ec2Client := ec2.New(session, conf)

	ec2Result, err := ec2Client.DescribeInstances(&ec2.DescribeInstancesInput{
		InstanceIds: ec2InstanceIDs,
	})
	common.FatalOnError(err)

	ec2ToIPs := map[string]string{}
	for _, reservation := range ec2Result.Reservations {
		for _, ec2Instance := range reservation.Instances {
			if ec2Instance.PublicIpAddress != nil {
				ec2ToIPs[*ec2Instance.InstanceId] = *ec2Instance.PublicIpAddress
			} else {
				ec2ToIPs[*ec2Instance.InstanceId] = *ec2Instance.PrivateIpAddress
			}
		}
	}

	for containerInstanceId, port := range bindings {
		ec2ID, ok := containerInstanceToEC2[containerInstanceId]
		if !ok {
			common.Fatalln("Could not resolve the container instance to EC2")
		}

		ip, ok := ec2ToIPs[ec2ID]
		if !ok {
			common.Fatalln("Could not get an IP address for EC2")
		}
		fmt.Printf("%s:%d\n", ip, port)
	}
}
