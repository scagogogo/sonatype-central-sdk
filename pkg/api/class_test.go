package api

import (
	"context"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestSearchByClassName(t *testing.T) {
	versionSlice, err := SearchByClassName(context.Background(), "guice", 10)
	assert.Nil(t, err)
	assert.True(t, len(versionSlice) > 0)
}
