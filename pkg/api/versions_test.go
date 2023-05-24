package api

import (
	"context"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestSearchByGroupIdAndArtifactId(t *testing.T) {
	versionSlice, err := ListVersions(context.Background(), "com.google.inject", "guice", -1)
	assert.Nil(t, err)
	assert.True(t, len(versionSlice) > 0)
}
