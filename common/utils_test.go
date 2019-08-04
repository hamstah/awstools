package common

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFlattenMapNested(t *testing.T) {
	m := map[string]interface{}{
		"a": map[string]interface{}{
			"b": map[string]interface{}{
				"c":   "abc",
				"d-e": "def",
			},
		},
	}

	flat, err := FlattenEnvVarMap(m)
	require.NoError(t, err)
	assert.Equal(t, "abc", flat["A_B_C"])
	assert.Equal(t, "def", flat["A_B_D_E"])
}
