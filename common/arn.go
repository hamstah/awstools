package common

import (
	"errors"
	"strings"
)

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
