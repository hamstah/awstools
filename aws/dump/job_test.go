package main

import (
	"fmt"
	"testing"

	"github.com/hamstah/awstools/common"
	"github.com/stretchr/testify/require"
)

func TestParseArn(t *testing.T) {
	t.Parallel()

	testCases := [][]string{
		[]string{"arn:partition:service:region:account-id:resource", "{partition service region account-id  resource }"},
		[]string{"arn:partition:service:region:account-id:resourcetype/resource", "{partition service region account-id resourcetype resource }"},
		[]string{"arn:partition:service:region:account-id:resourcetype/resource/qualifier", "{partition service region account-id resourcetype resource qualifier}"},
		[]string{"arn:partition:service:region:account-id:resourcetype:resource", "{partition service region account-id resourcetype resource }"},
		[]string{"arn:partition:service:region:account-id:resourcetype/resource:qualifier", "{partition service region account-id resourcetype resource qualifier}"},
		[]string{"arn:partition:service:region:account-id:resourcetype:resource:qualifier", "{partition service region account-id resourcetype resource qualifier}"},
	}
	for _, testCase := range testCases {
		parsed, err := common.ParseARN(testCase[0])
		require.NoError(t, err)
		require.NotNil(t, parsed)
		require.Equal(t, testCase[1], fmt.Sprintf("%s", *parsed))
	}

}
