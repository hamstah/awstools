package main

import (
	"time"

	"github.com/aws/aws-sdk-go/service/ecs"
)

type Account struct {
	AccountName string `json:"account_name"`
	Role        string `json:"role"`
	ExternalID  string `json:"external_id"`
	Prefix      string `json:"prefix"`
}

type Config struct {
	Accounts []Account `json:"accounts"`
}

type AccountState struct {
	Account         *Account                       `json:"account"`
	UpdatedAt       time.Time                      `json:"updated_at"`
	Clusters        []*ecs.Cluster                 `json:"clusters"`
	Services        []*ecs.Service                 `json:"services"`
	TaskDefinitions map[string]*ecs.TaskDefinition `json:"task_definitions"`
}

type Message struct {
	Message string `json:"message"`
	Type    string `json:"type"`
}
