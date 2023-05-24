package api

import (
	"context"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestSearchByArtifactId(t *testing.T) {
	artifactSlice, err := SearchByArtifactId(context.Background(), "guice", -1)
	assert.Nil(t, err)
	assert.True(t, len(artifactSlice) > 0)
}
