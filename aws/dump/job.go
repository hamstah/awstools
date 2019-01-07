package main

import (
	"errors"
	"strings"
)

type Resource struct {
	ID        string            `json:"id"`
	ARN       string            `json:"arn"`
	Service   string            `json:"service"`
	Type      string            `json:"type"`
	AccountID string            `json:"account_id"`
	Region    string            `json:"region"`
	Metadata  map[string]string `json:"metadata"`
	ManagedBy map[string]string `json:"managed_by"`
}

func (r *Resource) UniqueID() string {
	if r.ARN == "" {
		return r.ID
	}

	return r.ARN
}

func NewResource(arn string) (*Resource, error) {
	parsed, err := ParseARN(arn)
	if err != nil {
		return nil, err
	}

	return &Resource{
		ID:        parsed.Resource,
		ARN:       arn,
		Service:   parsed.Service,
		Type:      parsed.ResourceType,
		AccountID: parsed.AccountID,
		Region:    parsed.Region,
		Metadata:  map[string]string{},
	}, nil
}

type FetchResult struct {
	Resources []Resource
	Error     error
}

type Fetcher func(*Session) *FetchResult

type Job struct {
	Fetcher Fetcher
	Session *Session
}

func worker(id int, jobs <-chan Job, results chan<- *FetchResult) {
	for job := range jobs {
		results <- job.Fetcher(job.Session)
	}
}

func Run(sessions []*Session, fetchers []Fetcher) []Resource {
	resources := []Resource{}

	count := len(sessions) * len(fetchers)

	jobs := make(chan Job, count)
	results := make(chan *FetchResult, count)

	for w := 0; w < 10; w++ {
		go worker(w, jobs, results)
	}

	for _, fetcher := range fetchers {
		for _, session := range sessions {
			jobs <- Job{
				Fetcher: fetcher,
				Session: session,
			}
		}
	}
	close(jobs)

	for i := 0; i < count; i++ {
		result := <-results
		if result.Error == nil {
			resources = append(resources, result.Resources...)
		}
	}
	return resources
}

/*
arn:partition:service:region:account-id:resource
arn:partition:service:region:account-id:resourcetype/resource
arn:partition:service:region:account-id:resourcetype/resource/qualifier

arn:partition:service:region:account-id:resourcetype:resource
arn:partition:service:region:account-id:resourcetype/resource:qualifier
arn:partition:service:region:account-id:resourcetype:resource:qualifier
*/

type ARN struct {
	Partition    string
	Service      string
	Region       string
	AccountID    string
	ResourceType string
	Resource     string
	Qualifier    string
}

func ParseARN(arn string) (*ARN, error) {
	parts := strings.Split(arn, ":")
	if len(parts) < 6 {
		return nil, errors.New("Invalid format")
	}

	result := &ARN{
		Partition: parts[1],
		Service:   parts[2],
		Region:    parts[3],
		AccountID: parts[4],
	}

	if len(parts) == 6 {
		/*
		   arn:partition:service:region:account-id:resource
		   arn:partition:service:region:account-id:resourcetype/resource
		   arn:partition:service:region:account-id:resourcetype/resource/qualifier
		*/

		resourceParts := strings.Split(parts[5], "/")

		if len(resourceParts) == 1 {
			result.Resource = resourceParts[0]
			return result, nil
		}

		result.ResourceType = resourceParts[0]
		result.Resource = resourceParts[1]

		if len(resourceParts) > 2 {
			result.Qualifier = resourceParts[2]
		}
		return result, nil
	}

	if len(parts) == 8 {
		result.ResourceType = parts[5]
		result.Resource = parts[6]
		result.Qualifier = parts[7]
		return result, nil
	}

	resourceParts := strings.Split(parts[5], "/")
	result.ResourceType = resourceParts[0]
	if len(resourceParts) == 1 {
		result.Resource = parts[6]
		return result, nil
	}
	result.Resource = resourceParts[1]
	result.Qualifier = parts[6]

	return result, nil
}
