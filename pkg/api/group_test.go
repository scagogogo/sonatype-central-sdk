package api

import (
	"context"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestSearchByGroupId(t *testing.T) {
	artifactSlice, err := SearchByGroupId(context.Background(), "com.google.inject", -1)
	assert.Nil(t, err)
	assert.True(t, len(artifactSlice) > 0)
}
