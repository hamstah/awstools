package common

import (
	"errors"
	"fmt"
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

func (a *ARN) String() string {

	resource := ""
	for _, part := range []string{
		a.ResourceType,
		a.Resource,
		a.Qualifier,
	} {
		if part == "" {
			continue
		}
		if len(resource) > 0 {
			resource += "/" + part
		} else {
			resource = part
		}
	}

	return strings.Join([]string{
		"arn",
		a.Partition,
		a.Service,
		a.Region,
		a.AccountID,
		resource,
	}, ":")
}

func ReplaceAccountID(arn, accountID string) (string, error) {
	parsed, err := ParseARN(arn)
	if err != nil {
		return "", fmt.Errorf("failed to parse ARN: %w", err)
	}

	parsed.AccountID = accountID
	return parsed.String(), nil
}

func ReplaceAccountIDPtr(arn *string, accountID string) (*string, error) {
	if arn == nil {
		return nil, nil
	}

	res, err := ReplaceAccountID(*arn, accountID)
	if err != nil {
		return nil, err
	}
	return &res, nil
}

func ParseARN(arn string) (*ARN, error) {
	parts := strings.Split(arn, ":")
	if len(parts) < 6 {
		return nil, errors.New("invalid format")
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
			result.Qualifier = strings.Join(resourceParts[2:], "/")
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
	result.Qualifier = strings.Join(parts[6:], "/")

	return result, nil
}
