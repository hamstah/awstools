package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIPRanges(t *testing.T) {

	inner := []string{
		"127.0.0.1/32",
		"172.16.0.0/24",
		"172.16.123.45/32",
	}

	outer := []string{
		"127.0.0.1/32",
		"172.16.0.0/16",
	}

	res, err := ValidateIPRanges(inner, outer)
	assert.NoError(t, err)
	assert.Equal(t, res, inner)
}

func TestIPRangesNoInner(t *testing.T) {

	inner := []string{}

	outer := []string{
		"127.0.0.1/32",
		"172.16.0.0/16",
	}

	res, err := ValidateIPRanges(inner, outer)
	assert.NoError(t, err)
	assert.Equal(t, res, outer)
}

func TestIpRangesNoOuter(t *testing.T) {
	inner := []string{
		"127.0.0.1/32",
		"172.16.0.0/24",
		"172.16.123.45/32",
	}

	outer := []string{}

	res, err := ValidateIPRanges(inner, outer)
	assert.NoError(t, err)
	assert.Equal(t, res, inner)
}

func TestIPRangesInvalid(t *testing.T) {

	inner := []string{
		"172.16.123.45/32",
	}

	outer := []string{
		"172.16.128.0/17",
	}

	res, err := ValidateIPRanges(inner, outer)
	assert.Error(t, err)
	assert.Nil(t, res)
}
