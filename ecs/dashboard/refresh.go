package main

import (
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go/service/ecs"
	log "github.com/sirupsen/logrus"
)

func fetchAccount(account Account) (*AccountState, error) {
	logger := log.WithFields(log.Fields{
		"account": account.AccountName,
	})
	conf := createAWSConfig(account.Role, account.ExternalID, region, awsSession)
	svc := ecs.New(awsSession, &conf)
	clusterArns, err := getClusters(svc)
	services := []*ecs.Service{}
	taskDefinitions := map[string]*ecs.TaskDefinition{}

	if err != nil {
		messages = append(messages, Message{
			Message: fmt.Sprintf("Could not list clusters for %s (%s)", account.AccountName, err),
			Type:    "error",
		})
		logger.Error(err)
		return nil, err
	}

	clusters, err := describeClusters(svc, clusterArns)
	if err != nil {
		logger.Error(err)
		return nil, err
	}
	for _, cluster := range clusters {
		serviceNames, err := listServices(svc, cluster.ClusterArn)
		if err != nil {
			logger.Error(err)
			continue
		}

		if len(serviceNames) == 0 {
			continue
		}

		newServices, err := describeServices(svc, cluster.ClusterName, serviceNames)
		if err != nil {
			logger.Error(err)
			continue
		}
		services = append(services, newServices...)
	}

	for _, service := range services {
		if taskDefinitions[*service.ServiceName] == nil {
			taskDefinition, err := getTaskDefinition(svc, service.TaskDefinition)
			if err != nil {
				logger.Error(err)
				continue
			}
			taskDefinitions[*service.TaskDefinition] = taskDefinition
		}
	}

	return &AccountState{
		Account:         &account,
		Clusters:        clusters,
		Services:        services,
		UpdatedAt:       time.Now(),
		TaskDefinitions: taskDefinitions,
	}, nil
}

func worker(id int, jobs <-chan Account, results chan<- *AccountState) {
	for account := range jobs {
		state, _ := fetchAccount(account)
		results <- state
	}
}

func updateAccounts() {
	messagesM.Lock()
	messages = []Message{}
	messagesM.Unlock()

	jobs := make(chan Account, 100)
	results := make(chan *AccountState, 100)

	for w := 1; w <= 3; w++ {
		go worker(w, jobs, results)
	}

	for _, account := range config.Accounts {
		jobs <- account
	}
	close(jobs)

	for a := 1; a <= len(config.Accounts); a++ {
		result := <-results
		if result != nil {
			accountStatesM.Lock()
			accountStates[result.Account.AccountName] = result
			accountStatesM.Unlock()
		}
	}
}
