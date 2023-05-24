package api

import (
	"context"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestSearchBySha1(t *testing.T) {
	versionSlice, err := SearchBySha1(context.Background(), "35379fb6526fd019f331542b4e9ae2e566c57933", -1)
	assert.Nil(t, err)
	assert.True(t, len(versionSlice) > 0)
}
