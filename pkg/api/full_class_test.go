package api

import (
	"context"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestSearchByFullyQualifiedClassName(t *testing.T) {
	versionSlice, err := SearchByFullyQualifiedClassName(context.Background(), "guice", 10)
	assert.Nil(t, err)
	assert.True(t, len(versionSlice) > 0)
}
