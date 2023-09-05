package common

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseARN(t *testing.T) {
	arn := "arn:aws:iam::330428913683:role/aws-service-role/autoscaling.amazonaws.com/AWSServiceRoleForAutoScaling"
	parsed, err := ParseARN(arn)
	require.NoError(t, err)

	assert.Equal(t, "role", parsed.ResourceType)
	assert.Equal(t, "aws-service-role", parsed.Resource)
	assert.Equal(t, "autoscaling.amazonaws.com/AWSServiceRoleForAutoScaling", parsed.Qualifier)
}

func TestReplaceAccountID(t *testing.T) {
	arn := "arn:aws:iam::330428913683:role/aws-service-role/autoscaling.amazonaws.com/AWSServiceRoleForAutoScaling"
	accountID := "000000999999"

	expectedArn := "arn:aws:iam::000000999999:role/aws-service-role/autoscaling.amazonaws.com/AWSServiceRoleForAutoScaling"
	replaced, err := ReplaceAccountID(arn, accountID)
	require.NoError(t, err)
	assert.Equal(t, expectedArn, replaced)
}
